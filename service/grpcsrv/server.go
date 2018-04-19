package grpcsrv

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"

	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"google.golang.org/grpc"
)

var (
	emptyRsp         = &grpccomm.CommResponse{}
	emptyHeader      = reflect.ValueOf(&grpccomm.CommHeader{})
	mcodePrefix      string
	defaultService   *Service
	defaultGroupName = "defaultGrp"
)

// Service 服务
type Service struct {
	groups      map[string]*InterfaceGroup
	methodIndex map[string]*InterfaceGroup
	proxyRules  []*RouteProxy

	proxies []*ProxyTarget
}

// RouteProxy 代理
type RouteProxy struct {
	group  *InterfaceGroup
	target *ProxyTarget
}

// Do 执行代理
func (routeProxy *RouteProxy) Do(ctx context.Context, req *grpccomm.CommRequest) (*grpccomm.CommResponse, bool) {
	ok, service, err := routeProxy.target.detect(ctx, req)
	if !ok {
		return nil, false
	}

	if err != nil {
		return newRspFromError(err), false
	}

	if err := routeProxy.group.doPipe(ctx, req); err != nil {
		return newRspFromError(err), true
	}

	return service.Unary().Context(ctx).CommRequest(req), true
}

func init() {
	defaultService = newService()
}

func newService() *Service {
	service := &Service{}
	service.Group(defaultGroupName)

	return service
}

// RpcRequest 处理请求
func (service *Service) RpcRequest(ctx context.Context, commReq *grpccomm.CommRequest) (*grpccomm.CommResponse, error) {
	exec := func(ctx context.Context, commReq *grpccomm.CommRequest) *grpccomm.CommResponse {
		if commReq.ReqInterface == "" {
			return newErrorRsp("GRPC_METHOD_NOTFOUND", "grpc method missing")
		}

		// 查找本地是否处理
		group, exist := service.methodIndex[commReq.ReqInterface]
		if exist {
			return group.RpcRequest(ctx, commReq)
		}

		// 在代理里面寻找处理
		for _, rule := range service.proxyRules {
			rsp, executed := rule.Do(ctx, commReq)
			if !executed {
				continue
			}

			return rsp
		}

		return newErrorRsp("GRPC_METHOD_NOTFOUND", "grpc %s not registered", commReq.ReqInterface)
	}

	var f = exec
	for i := range usingHooks {
		f = usingHooks[len(usingHooks)-1-i](f)
	}

	return f(ctx, commReq), nil
}

// regProxy 注册一条代理规则
func (service *Service) regProxy(group *InterfaceGroup, t *ProxyTarget) {
	service.proxyRules = append(service.proxyRules, &RouteProxy{
		group:  group,
		target: t,
	})
}

// regGroup 注册一个接口
func (service *Service) regInterface(g *InterfaceGroup, reqName string) error {
	if service.methodIndex == nil {
		service.methodIndex = make(map[string]*InterfaceGroup, 1)
	}
	_, exist := service.methodIndex[reqName]
	if exist {
		unexpectError("duplicate register %v", reqName)
	}

	service.methodIndex[reqName] = g

	return nil
}

// Group 获取一个规则分组，如果不存在，则创建一个
func (service *Service) Group(name string, pipes ...RequestPipeFunc) *InterfaceGroup {
	group, isNew := service.newGroup(name, pipes...)
	if isNew {
		group.parents = []*InterfaceGroup{}
	}

	return group
}

func (service *Service) newGroup(name string, pipes ...RequestPipeFunc) (*InterfaceGroup, bool) {
	if service.groups == nil {
		service.groups = make(map[string]*InterfaceGroup, 1)
	}

	group, exist := service.groups[name]
	if !exist {
		stdGroup := &InterfaceGroup{
			name:    name,
			infos:   map[string]*RegisterInfo{},
			service: service,
		}

		service.groups[name] = stdGroup
		group = stdGroup
	}

	if len(pipes) > 0 {
		group.Use(pipes...)
	}

	return group, !exist
}

// Run 启动服务
func (service *Service) Run(host, errPrefix string, grpcOpts ...grpc.ServerOption) error {
	return ListenAndServe(host, errPrefix, grpcOpts...)
}

// ServeHTTP 用于HTTP调试
func (service *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service.debugHttp(w, r)
}

// DebugHttpHandler 返回HTTP调试
func (service *Service) DebugHttpHandler() http.Handler {
	return service
}

// Register 绑定函数
func (service *Service) Register(reqBody interface{}, f interface{}) {
	service.groups[defaultGroupName].Register(reqBody, f)
}

// ProxyService 反向代理服务
func (service *Service) ProxyService(serviceName string, target interface{}, reqTypes ...interface{}) {
	service.groups[defaultGroupName].ProxyService(serviceName, target, reqTypes...)
}

// ProxyInterface 反向代理接口
func (service *Service) ProxyInterface(req interface{}, target interface{}) {
	service.groups[defaultGroupName].ProxyInterface(req, target)
}

// RequestMethod 请求处理
type RequestMethod struct {
	f func(ctx context.Context, commReq *grpccomm.CommRequest) *grpccomm.CommResponse
}

// ListenAndServe GRPC监听
func ListenAndServe(host, errPrefix string, grpcOpts ...grpc.ServerOption) error {
	mcodePrefix = errPrefix

	lis, err := net.Listen("tcp", host)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(grpcOpts...)

	grpccomm.RegisterCommServiceServer(grpcServer, defaultService)

	return grpcServer.Serve(lis)
}

// RegisterInfo 注册信息
type RegisterInfo struct {
	ctxInIndex    int // ctx 位于函数输入的位置,-1 表示无此输入参数
	reqInIndex    int // req 位于函数输入的位置,-1 表示无此输入参数
	headerInIndex int // header 位于函数输入的位置, -1 表示无此输入参数
	commReqIndex  int // commReq 位于函数输入的位置，-1 表示无此输入参数
	inNum         int

	rspOutIndex     int // rsp 位于函数输出的位置，-1 表示无输此输出参数
	errIndex        int // err 位于函数输出的位置，-1 表示无此输出参数
	commRspOutIndex int // commRsp 位于函数输入位置，-1 表示无此输出参数
	outNum          int

	newBody      func() reflect.Value
	newHeader    func() reflect.Value
	call         func(in []reflect.Value) []reflect.Value
	callFuncName string
	reqName      string
}

// PipeLineFunc 管道函数
type PipeLineFunc func(*grpccomm.CommRequest) error

// Register 绑定函数
func Register(reqBody interface{}, f interface{}) {
	defaultService.Register(reqBody, f)
}

// ProxyService 反向代理服务
func ProxyService(serviceName string, target interface{}, reqTypes ...interface{}) {
	defaultService.ProxyService(serviceName, target, reqTypes...)
}

// ProxyInterface 反向代理接口
func ProxyInterface(req interface{}, target interface{}) {
	defaultService.ProxyInterface(req, target)
}

// Group 返回指定名称的接口组
func Group(name string, pipes ...RequestPipeFunc) *InterfaceGroup {
	return defaultService.Group(name, pipes...)
}

// Run 启动默认服务
func Run(host, errPrefix string, grpcOpts ...grpc.ServerOption) error {
	return defaultService.Run(host, errPrefix, grpcOpts...)
}

// DebugHttpHandler 返回GRPC调试处理
func DebugHttpHandler() http.Handler {
	return defaultService.DebugHttpHandler()
}

// New  新建服务
func New() *Service {
	return newService()
}
