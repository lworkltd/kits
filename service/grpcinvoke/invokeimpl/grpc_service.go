package invokeimpl

import (
	"fmt"
	"sync"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/lworkltd/kits/service/grpcinvoke"
	grpc "google.golang.org/grpc"
)

// GrpcConnBalancer 负责多个GrpcConn对象的均衡
// 为多个GRPC连接进行负载均衡，同时按需增加GRPC的连接
type GrpcConnBalancer struct {
	conns       []*grpc.ClientConn
	index       int
	mutex       sync.RWMutex
	maxConns    int
	discovery   grpcinvoke.DiscoveryFunc
	useCnt      int
	lastCntTime time.Time
	target      string
	done        bool
}

// newGrpcConnBalancer 为多个GRPC连接进行负载均衡，同时按需增加GRPC的连接
func newGrpcConnBalancer(target string, maxConns int, discovery grpcinvoke.DiscoveryFunc) *GrpcConnBalancer {
	return &GrpcConnBalancer{
		maxConns:  maxConns,
		target:    target,
		discovery: discovery,
	}
}

// Get 获取一个连接
func (grpcConnBalancer *GrpcConnBalancer) Get() (*grpc.ClientConn, error) {
	grpcConnBalancer.mutex.Lock()
	defer grpcConnBalancer.mutex.Unlock()
	if grpcConnBalancer.done {
		return nil, fmt.Errorf("closed")
	}

	increase := false
	totalConns := len(grpcConnBalancer.conns)
	if totalConns < grpcConnBalancer.maxConns {
		// 首次创建
		if totalConns == 0 {
			increase = true
		}

		// 如果一秒以内并发超过所有连接能够承载的量，则新增连接
		if time.Now().Sub(grpcConnBalancer.lastCntTime) > time.Second {
			if grpcConnBalancer.useCnt > 2000*totalConns {
				increase = true
			}
			grpcConnBalancer.lastCntTime = time.Now()
		}
	}

	if increase {
		conn, err := dialGrpcConnByDiscovery(grpcConnBalancer.target, grpcConnBalancer.discovery)
		if err != nil {
			return nil, err
		}
		grpcConnBalancer.conns = append(grpcConnBalancer.conns, conn)
		totalConns++
	}

	if grpcConnBalancer.index > totalConns {
		grpcConnBalancer.index = 0
	}

	conn := grpcConnBalancer.conns[0]
	grpcConnBalancer.index++

	return conn, nil
}

// Close 关闭
func (grpcConnBalancer *GrpcConnBalancer) Close() error {
	grpcConnBalancer.mutex.Lock()
	defer grpcConnBalancer.mutex.Unlock()
	if grpcConnBalancer.done {
		return fmt.Errorf("duplicate close")
	}
	for _, conn := range grpcConnBalancer.conns {
		conn.Close()
	}
	grpcConnBalancer.conns = nil
	grpcConnBalancer.done = true

	return nil
}

// grpcService 是GRPC调用方的代理实现
type grpcService struct {
	name              string
	freeConnAfterUsed bool
	mutex             sync.RWMutex
	lastConnUsedTime  time.Time
	conn              *grpc.ClientConn

	useTracing  bool
	useCircuit  bool
	doLogger    bool
	hystrixInfo hystrix.CommandConfig

	connLb *GrpcConnBalancer

	remove func()

	done bool
}

// Grpc 返回调用实例
func (grpcService *grpcService) Unary(args ...string) grpcinvoke.Client {
	grpcService.mutex.Lock()
	if grpcService.done {
		grpcService.mutex.Unlock()
		return newErrorGrpcClient(fmt.Errorf("invoke service closed"))
	}
	grpcService.mutex.Unlock()

	conn, err := grpcService.connLb.Get()
	if err != nil {
		return newErrorGrpcClient(err)
	}

	var callName string
	if len(args) > 0 {
		callName = args[0]
	}

	return grpcService.newGrpcClient(callName, conn)
}

func (grpcService *grpcService) newGrpcClient(callName string, conn *grpc.ClientConn) *grpcClient {
	return &grpcClient{
		callName:          callName,
		conn:              conn,
		freeConnAfterUsed: grpcService.freeConnAfterUsed,
		serviceName:       grpcService.name,
		since:             time.Now().UTC(),
		useCircuit:        grpcService.useCircuit,
		hystrixInfo:       grpcService.hystrixInfo,
		useTracing:        grpcService.useTracing,
		doLogger:          grpcService.doLogger,
	}
}

func (grpcService *grpcService) Close() error {
	grpcService.mutex.Lock()
	defer grpcService.mutex.Unlock()
	if grpcService.done == true {
		return fmt.Errorf("closed aready")
	}
	grpcService.done = true

	if grpcService.remove != nil {
		grpcService.remove()
	}

	grpcService.connLb.Close()

	grpcService.done = true

	return nil
}

func (grpcService *grpcService) Name() string {
	return grpcService.name
}

// UseCircuit 启用熔断
func (grpcService *grpcService) UseCircuit(enable bool) grpcinvoke.Service {
	grpcService.useCircuit = enable
	return grpcService
}

// MaxConcurrent 最大并发请求
func (grpcService *grpcService) MaxConcurrent(maxConn int) grpcinvoke.Service {
	if maxConn < 30 {
		maxConn = 30
	} else if maxConn > 10000 {
		maxConn = 10000
	}
	grpcService.hystrixInfo.MaxConcurrentRequests = maxConn

	return grpcService
}

// Timeout 请求超时立即返回时间
func (grpcService *grpcService) Timeout(timeout time.Duration) grpcinvoke.Service {
	if timeout < 10*time.Millisecond {
		timeout = time.Millisecond * 10
	} else if timeout > 10000*time.Millisecond {
		timeout = 10000 * time.Millisecond
	}

	grpcService.hystrixInfo.Timeout = int(timeout / time.Millisecond)

	return grpcService
}

// PercentThreshold 最大错误容限
func (grpcService *grpcService) PercentThreshold(thresholdPercent int) grpcinvoke.Service {
	if thresholdPercent < 5 {
		thresholdPercent = 5
	} else if thresholdPercent > 100 {
		thresholdPercent = 100
	}

	grpcService.hystrixInfo.RequestVolumeThreshold = thresholdPercent

	return grpcService
}
