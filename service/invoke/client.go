package invoke

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/monitor"
	"github.com/opentracing/opentracing-go"
)

const (
	HTTP_HEADER_CONTENT_TYPE = "Content-Type"

	HTTP_HEADER_CONTENT_TYPE_JSON = "application/json"
)

type client struct {
	service      Service
	path         string
	createTime   time.Time
	errInProcess error

	method        string
	host          string
	scheme        string
	serverid      string
	circuitConfig hystrix.CommandConfig

	headers map[string]string
	queries map[string][]string
	routes  map[string]string
	payload func() ([]byte, error)

	logFields  map[string]interface{}
	ctx        context.Context
	useTracing bool
	useCircuit bool
	fallback   func(error) error
}

func (client *client) circuitName() string {
	return client.serverid
}

//未设置hytrix参数，或者参数不合理，使用默认熔断策略
func (client *client) hytrixCommand() string {
	return client.serverid + client.method + client.path
}

func (client *client) tracingName() string {
	return client.serverid
}

func (client *client) clear() {
	client.service = nil
	client.path = ""
	client.createTime = time.Unix(0, 0)
	client.errInProcess = nil
	client.headers = nil
	client.routes = nil
	client.queries = nil
	client.method = "GET"
	client.host = ""
	client.scheme = "http"
	client.payload = nil
	client.logFields = make(map[string]interface{}, 10)
	client.ctx = nil
}

func (client *client) Tls() Client {
	if client.errInProcess != nil {
		return client
	}

	client.scheme = "https"

	return client
}

func (client *client) Fallback(func(error) error) Client {
	if client.errInProcess != nil {
		return client
	}
	return client
}

func (client *client) Header(headerName, headerValue string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.headers == nil {
		client.headers = map[string]string{headerName: headerValue}
		return client
	}

	client.headers[headerName] = headerValue

	return client
}

func (client *client) Headers(headers map[string]string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.headers == nil {
		client.headers = make(map[string]string, len(headers))
	}

	for key, value := range headers {
		client.headers[key] = value
	}

	return client
}

func (client *client) Query(queryName, queryValue string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.queries == nil {
		client.queries = map[string][]string{
			queryName: []string{queryValue},
		}
		return client
	}

	queries := client.queries[queryName]
	queries = append(queries, queryValue)
	client.queries[queryName] = queries

	return client
}

func (client *client) QueryArray(queryName string, queryValues ...string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.queries == nil {
		client.queries = map[string][]string{
			queryName: queryValues,
		}
		return client
	}

	queries := client.queries[queryName]
	queries = append(queries, queryValues...)
	client.queries[queryName] = queries

	return client
}

func (client *client) Queries(queryValues map[string][]string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.queries == nil {
		client.queries = make(map[string][]string, len(queryValues))
		return client
	}

	for key, queryValueSlice := range queryValues {
		queries := client.queries[key]
		queries = append(queries, queryValueSlice...)
		client.queries[key] = queries
	}

	return client
}

func (client *client) Route(routeName, routeTo string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.routes == nil {
		client.routes = map[string]string{routeName: routeTo}
		return client
	}

	client.routes[routeName] = routeTo

	return client
}

func (client *client) Routes(routes map[string]string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.routes == nil {
		client.routes = make(map[string]string, len(routes))
	}

	for routeName, route := range routes {
		client.routes[routeName] = route
	}

	return client
}

func (client *client) Json(payload interface{}) Client {
	if client.errInProcess != nil {
		return client
	}

	client.payload = func() ([]byte, error) {
		return json.Marshal(payload)
	}

	return client
}

func (client *client) Proto(payload proto.Message) Client {
	if client.errInProcess != nil {
		return client
	}

	client.payload = func() ([]byte, error) {
		return proto.Marshal(payload)
	}

	return client
}

func (client *client) Body(payload []byte) Client {
	if client.errInProcess != nil {
		return client
	}

	client.payload = func() ([]byte, error) {
		return payload, nil
	}
	return client
}

func (client *client) Context(ctx context.Context) Client {
	if client.errInProcess != nil {
		return client
	}

	client.ctx = ctx

	return client
}

func (client *client) Hystrix(timeOutMillisecond, maxConn, thresholdPercent int) Client {
	if client.errInProcess != nil {
		return client
	}
	//设置值不合理时调整
	if timeOutMillisecond < 10 {
		timeOutMillisecond = 10
	} else if timeOutMillisecond > 10000 {
		timeOutMillisecond = 10000
	}
	if maxConn < 30 {
		maxConn = 30
	} else if maxConn > 10000 {
		maxConn = 10000
	}
	if thresholdPercent < 5 {
		thresholdPercent = 5
	} else if thresholdPercent > 100 {
		thresholdPercent = 100
	}

	client.circuitConfig.Timeout = timeOutMillisecond
	client.circuitConfig.MaxConcurrentRequests = maxConn
	client.circuitConfig.ErrorPercentThreshold = thresholdPercent
	return client
}

func (client *client) Exec(out interface{}) (int, error) {
	if client.useTracing {
		span, ctx := opentracing.StartSpanFromContext(client.ctx, client.tracingName())
		client.ctx = ctx
		defer span.Finish()
	}

	beginTime := time.Now()
	var err error
	var status int
	if !client.useCircuit {
		status, err = client.exec(out, nil)
	} else {
		client.updateHystrix()

		var cancel context.CancelFunc
		err = hystrix.Do(client.hytrixCommand(), func() error {
			s, err := client.exec(out, &cancel)
			status = s
			return err
		}, client.fallback)
		if nil != err && nil != cancel {
			cancel()
		}
	}
	client.processExecMonitorReport(status, err, beginTime)

	if doLogger {
		fileds := logrus.Fields{
			"service":    client.service.Name(),
			"service_id": client.serverid,
			"method":     client.method,
			"path":       client.path,
			"endpoint":   client.host,
			"cost":       time.Since(client.createTime),
		}

		if doLoggerParam {
			fileds["headers"] = client.headers
			fileds["queries"] = client.queries
			fileds["routes"] = client.routes
			fileds["payload"] = client.payload
		}

		if err != nil {
			logrus.WithFields(fileds).WithError(err).Error("Invoke service failed")
		} else {
			logrus.WithFields(fileds).Error("Invoke service done")
		}
	}

	return status, err
}

func (client *client) build() (*http.Request, error) {
	host, serviceNode, err := client.service.Remote()
	if err != nil {
		return nil, fmt.Errorf("discovery failed,%v", err)
	}

	client.host, client.serverid = host, serviceNode

	path, err := parsePath(client.path, client.routes)
	if err != nil {
		client.logFields["error"] = "routes parameter invalid"
		client.logFields["routes"] = client.routes
		return nil, err
	}

	if client.host == "" {
		client.logFields["error"] = "no avaliable remote"
		return nil, fmt.Errorf("remote is emtpy")
	}

	url, err := makeUrl(client.scheme, client.host, path, client.queries)

	if err != nil {
		client.logFields["scheme"] = client.scheme
		client.logFields["remote"] = client.host
		client.logFields["path"] = url
		client.logFields["queries"] = client.queries
		return nil, err
	}

	reader := &bytes.Reader{}
	if client.payload != nil {
		b, err := client.payload()
		if err != nil {
			return nil, err
		}
		client.logFields["payload"] = string(b)
		reader = bytes.NewReader(b)
	}

	request, err := http.NewRequest(client.method, url, reader)
	if err != nil {
		client.logFields["error"] = err
		return nil, fmt.Errorf("create http request failed,%v", err)
	}

	if client.ctx != nil {
		request = request.WithContext(client.ctx)
	}

	if _, ok := client.headers[HTTP_HEADER_CONTENT_TYPE]; !ok {
		request.Header.Add(HTTP_HEADER_CONTENT_TYPE, HTTP_HEADER_CONTENT_TYPE_JSON)
	}

	for headerKey, headerValue := range client.headers {
		request.Header.Add(headerKey, headerValue)
	}

	return request, nil
}

func (client *client) reportErrorToMonitor(code string, beginTime time.Time) {
	infc := "ACTIVE_" + client.method + "_" + client.path //ACTIVE表示主调
	//请求失败，上报失败计数和失败平均耗时
	timeNow := time.Now()
	var failedCountReport monitor.ReqFailedCountDimension
	failedCountReport.SName = monitor.GetCurrentServerName()
	failedCountReport.TName = client.service.Name()
	failedCountReport.TIP = client.host
	failedCountReport.Code = code
	failedCountReport.Infc = infc
	monitor.ReportReqFailed(&failedCountReport)

	var failedAvgTimeReport monitor.ReqFailedAvgTimeDimension
	failedAvgTimeReport.SName = monitor.GetCurrentServerName()
	failedAvgTimeReport.SIP = monitor.GetCurrentServerIP()
	failedAvgTimeReport.TName = client.service.Name()
	failedAvgTimeReport.TIP = client.host
	failedAvgTimeReport.Infc = infc
	monitor.ReportFailedAvgTime(&failedAvgTimeReport, (timeNow.UnixNano()-beginTime.UnixNano())/1e3) //耗时单位为微秒
}

func (client *client) reportSuccessToMonitor(beginTime time.Time) {
	infc := "ACTIVE_" + client.method + "_" + client.path //ACTIVE表示主调
	//请求失败，上报失败计数和失败平均耗时
	timeNow := time.Now()
	var succCountReport monitor.ReqSuccessCountDimension
	succCountReport.SName = monitor.GetCurrentServerName()
	succCountReport.SIP = monitor.GetCurrentServerIP()
	succCountReport.TName = client.service.Name()
	succCountReport.TIP = client.host
	succCountReport.Infc = infc
	monitor.ReportReqSuccess(&succCountReport)

	var succAvgTimeReport monitor.ReqSuccessAvgTimeDimension
	succAvgTimeReport.SName = monitor.GetCurrentServerName()
	succAvgTimeReport.SIP = monitor.GetCurrentServerIP()
	succAvgTimeReport.TName = client.service.Name()
	succAvgTimeReport.TIP = client.host
	succAvgTimeReport.Infc = infc
	monitor.ReportSuccessAvgTime(&succAvgTimeReport, (timeNow.UnixNano()-beginTime.UnixNano())/1e3) //耗时单位为微秒
}

//处理Response函数的http请求结果monitor上报
func (client *client) processResponseMonitorReport(resp *http.Response, beginTime time.Time) {
	if monitor.EnableReportMonitor() == false {
		return
	}

	if nil == resp {
		//请求失败，上报失败计数和失败平均耗时
		client.reportErrorToMonitor("-1", beginTime) //code暂时取"-1"
	} else {
		//把beginTime，infc，TName放入resp的header中，由ExtractHttpResponse取上报失败或成功
		infc := "ACTIVE_" + client.method + "_" + client.path //ACTIVE表示主调
		resp.Header.Set("Infc", infc)
		resp.Header.Set("TName", client.service.Name())
		resp.Header.Set("Endpoint", client.host) //请求的IP:Port，或者一个domain:Port/domain
		resp.Header.Set("BeginTime", strconv.FormatInt(beginTime.UnixNano()/1e3, 10))
	}
}

//处理Exec函数http请求结果的monitor上报
func (client *client) processExecMonitorReport(code int, err error, beginTime time.Time) {
	if monitor.EnableReportMonitor() == false {
		return
	}

	if nil != err {
		//请求失败，上报失败计数和失败平均耗时
		client.reportErrorToMonitor(strconv.FormatInt(int64(code), 10), beginTime)
	} else {
		//请求成功，上报成功计数和成功平均耗时
		client.reportSuccessToMonitor(beginTime)
	}
}

func (client *client) exec(out interface{}, cancel *context.CancelFunc) (int, error) {
	if client.errInProcess != nil {
		return 0, client.errInProcess
	}

	if nil != cancel && nil != client.ctx {
		client.ctx, *cancel = context.WithCancel(client.ctx)
	}
	request, err := client.build()
	if err != nil {
		return 0, err
	}

	cli := &http.Client{}
	resp, err := cli.Do(request)
	if err != nil {
		client.logFields["error"] = err
		return 0, err
	}

	client.logFields["status"] = resp.StatusCode
	client.logFields["status_code"] = resp.Status

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		client.logFields["error"] = "response status error"
		return resp.StatusCode, fmt.Errorf("reponse with bad status,%d", resp.StatusCode)
	}

	rsp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		client.logFields["error"] = err
		return resp.StatusCode, fmt.Errorf("Read response body failed")
	}
	defer resp.Body.Close()

	client.logFields["response_payload_len"] = len(rsp)

	err = json.Unmarshal(rsp, out)
	if err != nil {
		client.logFields["error"] = err
		client.logFields["content"] = string(cutBytes(rsp, 4096))
		return 0, fmt.Errorf("marshal result body failed")
	}

	return resp.StatusCode, nil
}

func (client *client) getResp(cancel *context.CancelFunc) (*http.Response, error) {
	if client.errInProcess != nil {
		return nil, client.errInProcess
	}

	if nil != cancel && nil != client.ctx {
		client.ctx, *cancel = context.WithCancel(client.ctx)
	}
	request, err := client.build()
	if err != nil {
		return nil, err
	}

	cli := &http.Client{}
	resp, err := cli.Do(request)
	if err != nil {
		client.logFields["error"] = err
		return nil, err
	}

	return resp, nil
}

func (client *client) updateHystrix() {
	hytrixCmd := client.hytrixCommand()
	//if _, exist, _ := hystrix.GetCircuit(hytrixCmd); exist {
	//	return
	//}

	hystrix.ConfigureCommand(hytrixCmd, client.circuitConfig)
}

func (client *client) Response() (*http.Response, error) {
	if client.useTracing {
		span, ctx := opentracing.StartSpanFromContext(client.ctx, client.tracingName())
		client.ctx = ctx
		defer span.Finish()
	}

	beginTime := time.Now()
	var err error
	var resp *http.Response
	if !client.useCircuit {
		resp, err = client.getResp(nil)
	} else {
		client.updateHystrix()

		var cancel context.CancelFunc
		err = hystrix.Do(client.hytrixCommand(), func() error {
			s, err := client.getResp(&cancel)
			resp = s
			return err
		}, client.fallback)
		if nil != err && nil != cancel {
			cancel() //cancel run client.getResp
		}
	}
	client.processResponseMonitorReport(resp, beginTime) //若resp为nil则上报错误，否则添加请求信息到header待进一步上报monitor数据

	if doLogger {
		fileds := logrus.Fields{
			"service":    client.service.Name(),
			"service_id": client.serverid,
			"method":     client.method,
			"path":       client.path,
			"endpoint":   client.host,
			"cost":       time.Since(client.createTime),
		}

		if doLoggerParam {
			fileds["headers"] = client.headers
			fileds["queries"] = client.queries
			fileds["routes"] = client.routes
			fileds["payload"] = client.payload
		}

		if err != nil {
			logrus.WithFields(fileds).WithError(err).Error("Invoke service failed")
		} else {
			logrus.WithFields(fileds).Info("Invoke service done")
		}
	}

	return resp, err

}
