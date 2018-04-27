package grpcinvoke

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
)

var doLogger = true

// DiscoveryFunc 服务发现的函数
type DiscoveryFunc func(name string) ([]string, []string, error)

type serviceDiscoveryFunc func() ([]string, []string, error)

// Option 用于初始化引擎的参数
type Option struct {
	Discover                     DiscoveryFunc
	LoadBalanceMode              string
	UseTracing                   bool
	UseCircuit                   bool
	DoLogger                     bool
	DefaultTimeout               time.Duration
	DefaultMaxConcurrentRequests int
	DefaultErrorPercentThreshold int
}

// Engine  引擎
type Engine interface {
	Init(*Option) error // 初始化
	Service(string) Service
	Addr(string) Service
}

// Service 服务代理
type Service interface {
	Unary(...string) Client
	Close() error
	Name() string

	Timeout(time.Duration) Service
	MaxConcurrent(int) Service
	PercentThreshold(int) Service
	UseCircuit(enable bool) Service
}

// Client 客户端
type Client interface {
	Body(proto.Message) Client
	ReqService(string) Client
	Context(context.Context) Client
	Header(proto.Message) Client
	Fallback(func(error) error) Client
	Timeout(time.Duration) Client
	MaxConcurrent(int) Client
	PercentThreshold(int) Client
	UseCircuit(enable bool) Client

	Response(proto.Message) code.Error
	CommRequest(*grpccomm.CommRequest) *grpccomm.CommResponse
}

// DefaultEngine 默认的引擎
var DefaultEngine Engine

// Init 初始化
func Init(option *Option) error {
	doLogger = option.DoLogger
	if true == option.UseCircuit {
		//未设置时的默认值
		if 0 == option.DefaultTimeout/time.Millisecond {
			option.DefaultTimeout = 1000 * time.Millisecond
		}
		if 0 == option.DefaultMaxConcurrentRequests {
			option.DefaultMaxConcurrentRequests = 200
		}
		if 0 == option.DefaultErrorPercentThreshold {
			option.DefaultErrorPercentThreshold = 20
		}

		//设置值不合理时调整
		if option.DefaultTimeout < 10*time.Millisecond {
			option.DefaultTimeout = 10 * time.Millisecond
		} else if option.DefaultTimeout > 10*time.Second {
			option.DefaultTimeout = 10 * time.Second
		}

		if option.DefaultMaxConcurrentRequests < 30 {
			option.DefaultMaxConcurrentRequests = 30
		} else if option.DefaultMaxConcurrentRequests > 10000 {
			option.DefaultMaxConcurrentRequests = 10000
		}

		if option.DefaultErrorPercentThreshold < 5 {
			option.DefaultErrorPercentThreshold = 5
		} else if option.DefaultErrorPercentThreshold > 100 {
			option.DefaultErrorPercentThreshold = 100
		}
	}
	return DefaultEngine.Init(option)
}

// Name 根据服务名称获取GRPC服务
func Name(name string) Service {
	return DefaultEngine.Service(name)
}

// Addr 根据服务地址获取GRPC服务
func Addr(addr string) Service {
	return DefaultEngine.Addr(addr)
}
