package urpcinvoke

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/monitor"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/lworkltd/kits/service/urpcsrv/urpccomm"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

type Client struct {
	err error

	callName    string
	body        proto.Message
	header      proto.Message
	fallback    func(error) error
	serviceId   string
	serviceName string
	reqService  string

	hystrixInfo hystrix.CommandConfig
	useTracing  bool
	useCircuit  bool
	doLogger    bool
	ctx         context.Context
	addr        string

	discovery func() (string, string, code.Error)

	since time.Time
}

func newErrorClient(err error) *Client {
	return &Client{err: err}
}

func (client *Client) Header(reqHeader proto.Message) *Client {
	if client.err != nil {
		return client
	}
	client.header = reqHeader
	return client
}

func (client *Client) ReqService(reqService string) *Client {
	if client.err != nil {
		return client
	}
	client.reqService = reqService
	return client
}

func (client *Client) Body(reqBody proto.Message) *Client {
	if client.err != nil {
		return client
	}

	if reqBody == nil {
		return client
	}

	client.body = reqBody

	if client.callName == "" {
		t := reflect.TypeOf(reqBody)
		client.callName = t.Elem().Name()
	}

	return client
}

func (client *Client) Fallback(f func(error) error) *Client {
	if client.err != nil {
		return client
	}
	client.fallback = f
	return client
}

func buildGrpcCommRequest(client *Client) (*urpccomm.CommRequest, code.Error) {
	in := &urpccomm.CommRequest{
		ReqInterface: client.callName,
	}

	if client.body != nil {
		bodyBytes, err := proto.Marshal(client.body)
		if err != nil {
			return nil, code.NewMcodef("BAD_REQUEST_BODY", "invaild body %v", err)
		}
		in.Body = bodyBytes
	}

	if client.header != nil {
		headerBytes, err := proto.Marshal(client.header)
		if err != nil {
			return nil, code.NewMcodef("BAD_REQUEST_HEADER", "invaild hader %v", err)
		}
		in.Header = headerBytes
	}

	reqService := client.reqService
	if reqService == "" {
		reqService = client.serviceName
	}
	in.ReqSercice = reqService

	if client.ctx == nil {
		client.ctx = context.Background()
	}

	return in, nil
}

func (client *Client) CommRequest(in *urpccomm.CommRequest) *urpccomm.CommResponse {
	if client.err != nil {
		return &urpccomm.CommResponse{
			Result:  false,
			Mcode:   "INVOKE_FAILED",
			Message: fmt.Sprintf("invoke %v failed,%v", client.serviceName, client.err),
		}
	}

	if client.ctx == nil {
		client.ctx = context.Background()
	}

	var (
		rsp *urpccomm.CommResponse
		err error
	)

	if !client.useCircuit {
		rsp, err = call0(client.ctx, client.addr, in)
	} else {
		client.updateHystrix()
		var cancel context.CancelFunc
		newCtx, cancel := context.WithCancel(client.ctx)
		err = hystrix.Do(client.hytrixCommand(), func() error {
			insideRsp, insideErr := call0(newCtx, client.addr, in)
			rsp = insideRsp
			return insideErr
		}, client.fallback)
		if nil != err && nil != cancel {
			cancel()
		}
	}

	if err != nil {
		return &urpccomm.CommResponse{
			Result:  false,
			Mcode:   "INVOKE_FAILED",
			Message: fmt.Sprintf("invoke %v failed,%v", client.serviceName, client.err),
		}
	}

	return rsp
}

func (client *Client) catchAndReturnError(originErr error) code.Error {
	if originErr == nil {
		return nil
	}
	var (
		cerr code.Error
	)

	if err, yes := originErr.(code.Error); yes {
		client.doLog(err)

		monitor.CommMonitorReport(
			err.Mcode(),
			monitor.GetCurrentServerName(),
			monitor.GetCurrentServerIP(),
			client.serviceName,
			"",
			fmt.Sprintf("ACTIVE_GRPC_%s", client.callName),
			client.since,
		)
		return err
	}

	cerr = code.NewMcode("UNKOWN_ERROR", originErr.Error())

	client.doLog(cerr)

	return cerr
}

// Context 指定上下文
func (client *Client) Context(ctx context.Context) *Client {
	if client.err != nil {
		return client
	}
	client.ctx = ctx
	return client
}

// Response 执行请求并获取结果
func (client *Client) Response(out proto.Message) code.Error {
	if client.err != nil {
		return client.catchAndReturnError(client.err)
	}
	addr, _, cerr := client.discovery()
	if cerr != nil {
		return client.catchAndReturnError(cerr)
	}
	client.addr = addr

	in, cerr := buildGrpcCommRequest(client)
	if cerr != nil {
		return client.catchAndReturnError(cerr)
	}

	var (
		rsp *urpccomm.CommResponse
		err error
	)

	if !client.useCircuit {
		rsp, err = call0(client.ctx, client.addr, in)
	} else {
		client.updateHystrix()
		var cancel context.CancelFunc
		newCtx, cancel := context.WithCancel(client.ctx)
		err = hystrix.Do(client.hytrixCommand(), func() error {
			insideRsp, insideErr := call0(newCtx, client.addr, in)
			rsp = insideRsp
			return insideErr
		}, client.fallback)
		if nil != err && nil != cancel {
			cancel()
		}
	}

	if err != nil {
		stat, ok := status.FromError(err)
		if ok {
			switch stat.Code {
			default:
				return client.catchAndReturnError(code.NewMcode("URPC_ERROR", err.Error()))
			}
		}

		if strings.Index(err.Error(), "hystrix: timeout") >= 0 {
			return client.catchAndReturnError(code.NewMcode("URPC_TIMEOUT", err.Error()))
		}

		return client.catchAndReturnError(code.NewMcode("URPC_ERROR", err.Error()))
	}

	if !rsp.Result {
		return client.catchAndReturnError(rsp.CodeError())
	}

	if out != nil {
		err := proto.Unmarshal(rsp.Body, out)
		if err != nil {
			return client.catchAndReturnError(code.NewMcode("GRPC_BAD_BODY", "bad grpc response body"))
		}
	}

	client.doLog(nil)

	monitor.CommMonitorReport(
		"",
		monitor.GetCurrentServerName(),
		monitor.GetCurrentServerIP(),
		client.serviceName,
		"",
		fmt.Sprintf("ACTIVE_GRPC_%s", client.callName),
		client.since,
	)

	return nil
}

func (client *Client) hytrixCommand() string {
	return fmt.Sprintf("%s/%s", client.serviceName, client.callName)
}

func (client *Client) updateHystrix() {
	if client.useCircuit {
		hytrixCmd := client.hytrixCommand()
		if _, exist, _ := hystrix.GetCircuit(hytrixCmd); exist {
			return
		}

		hystrix.ConfigureCommand(hytrixCmd, client.hystrixInfo)
	}
}

// UseCircuit 启用熔断
func (client *Client) UseCircuit(enable bool) *Client {
	if client.err != nil {
		return client
	}

	client.useCircuit = enable
	return client
}

func (client *Client) doLog(err code.Error) {
	if !client.doLogger {
		return
	}

	cost := time.Now().Sub(client.since)
	log := logrus.WithFields(logrus.Fields{
		"reqName":    client.callName,
		"reqService": client.serviceName,
		"latency":    cost,
	})

	if logrus.StandardLogger().Level >= logrus.DebugLevel {
		if client.body != nil {
			log = log.WithFields(logrus.Fields{
				"body": client.body,
			})
		}

		if client.header != nil {
			log = log.WithFields(logrus.Fields{
				"header": client.header,
			})
		}
	}

	if err != nil {
		log = log.WithFields(logrus.Fields{
			"error": err.Error(),
		})
		log.Error("GRPC INVOKE FAILED")
		return
	}

	log.Info("GRPC INVOKE DONE")
}

// MaxConcurrent 最大并发请求
func (client *Client) MaxConcurrent(maxConn int) *Client {
	if client.err != nil {
		return client
	}

	if maxConn < 30 {
		maxConn = 30
	} else if maxConn > 10000 {
		maxConn = 10000
	}
	client.hystrixInfo.MaxConcurrentRequests = maxConn
	return client
}

// Timeout 请求超时立即返回时间
func (client *Client) Timeout(timeout time.Duration) *Client {
	if client.err != nil {
		return client
	}

	if timeout < 10*time.Millisecond {
		timeout = time.Millisecond * 10
	} else if timeout > 10000*time.Millisecond {
		timeout = 10000 * time.Millisecond
	}

	client.hystrixInfo.Timeout = int(timeout / time.Millisecond)

	return client
}

// PercentThreshold 最大错误容限
func (client *Client) PercentThreshold(thresholdPercent int) *Client {
	if client.err != nil {
		return client
	}

	if thresholdPercent < 5 {
		thresholdPercent = 5
	} else if thresholdPercent > 100 {
		thresholdPercent = 100
	}

	client.hystrixInfo.RequestVolumeThreshold = thresholdPercent

	return client
}

// DoLogger 打印日志
func (client *Client) DoLogger(doLogger bool) *Client {
	client.doLogger = true
	return client
}

func call0(ctx context.Context, addr string, in *urpccomm.CommRequest) (*urpccomm.CommResponse, error) {
	var (
		err error
		rsp *urpccomm.CommResponse
	)
	finished := make(chan bool)
	go func() {
		defer close(finished)
		rsp, err = func() (*urpccomm.CommResponse, error) {
			conn, err := net.Dial("udp", addr)
			if err != nil {
				return nil, code.NewMcodef("BAD_REMOTE_ADDR", "%s is not a valid address", addr)
			}
			b, _ := proto.Marshal(in)
			n, err := conn.Write(b)
			if n != len(b) {
				return nil, code.NewMcodef("NETWORK_PROBLEM", "write error %v", addr)
			}

			rspBytes := make([]byte, 4098)
			conn.SetReadDeadline(time.Now().Add(time.Minute))
			rspN, err := conn.Read(rspBytes)
			if err != nil {
				return nil, code.NewMcodef("NETWORK_PROBLEM", "read error %v", addr)
			}
			rsp := &urpccomm.CommResponse{}
			err = proto.Unmarshal(rspBytes[:rspN], rsp)
			if err != nil {
				return nil, code.NewMcodef("BAD_PROTOCOL", "protocol error,%v", err)
			}

			return rsp, nil
		}()
	}()

	select {
	case <-ctx.Done():
		return nil, code.NewMcodef("TIME_OUT", "timeout")
	case <-finished:
		return rsp, err
	}
}
