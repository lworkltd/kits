package urpcsrv

import (
	"net"
	"net/http"
	"reflect"
	"syscall"

	context "golang.org/x/net/context"

	"github.com/lworkltd/kits/service/urpcsrv/urpccomm"
	"github.com/lworkltd/kits/service/version"
	"google.golang.org/grpc"
)

var (
	emptyRsp         = &urpccomm.CommResponse{}
	emptyHeader      = reflect.ValueOf(&urpccomm.CommHeader{})
	mcodePrefix      string
	defaultService   *Service
	defaultGroupName = "defaultGrp"
)

// Service 服务
type Service struct {
	server *grpc.Server

	groups      map[string]*InterfaceGroup
	methodIndex map[string]*InterfaceGroup
	hooks       []HookFunc
	proxies     []*ProxyTarget
}

func init() {
	defaultService = newService()
}

func newService() *Service {
	service := &Service{}
	service.Group(defaultGroupName)
	service.registerVersion()

	return service
}

// RpcRequest 处理请求
func (service *Service) RpcRequest(ctx context.Context, commReq *urpccomm.CommRequest) (*urpccomm.CommResponse, error) {
	exec := func(ctx context.Context, commReq *urpccomm.CommRequest) *urpccomm.CommResponse {
		if commReq.ReqInterface == "" {
			return newErrorRsp("GRPC_METHOD_NOTFOUND", "grpc method missing")
		}

		// 查找本地是否处理
		group, exist := service.methodIndex[commReq.ReqInterface]
		if exist {
			return group.RpcRequest(ctx, commReq)
		}

		return newErrorRsp("GRPC_METHOD_NOTFOUND", "grpc %s not registered", commReq.ReqInterface)
	}

	var f = exec
	for i := range service.hooks {
		f = service.hooks[len(service.hooks)-1-i](f)
	}

	return f(ctx, commReq), nil
}

// UseHook 增加Hook处理
func (service *Service) UseHook(hooks ...HookFunc) {
	service.hooks = append(service.hooks, hooks...)
}

func (service *Service) registerVersion() {
	service.Register("_AppVersion", func() (*version.VersionResponse, error) {
		return version.GetVersionInfo(), nil
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
func (service *Service) Run(addr string) error {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}

	defer conn.Close()
	for {
		buf := newBuf()
		n, fromAddr, err := conn.ReadFromUDP(buf)
		if err == syscall.EINVAL {
			return err
		}

		go func() {
			var commRsp *urpccomm.CommResponse

			dec := decoder{buf: buf[:n]}
			req, err := dec.Decode()
			if err != nil {
				commRsp = newErrorRsp("URPC_BAD_BODY", "bad data")
			} else {
				commRsp, _ = service.RpcRequest(context.Background(), req)
			}

			enc := encoder{rsp: commRsp}
			rb, _ := enc.Encode()

			conn.WriteToUDP(rb, fromAddr)
		}()
	}

	return nil
}

// Stop 停止服务
func (service *Service) Stop() {
	s := service.server
	if s == nil {
		return
	}

	s.Stop()
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

// RequestMethod 请求处理
type RequestMethod struct {
	f func(ctx context.Context, commReq *urpccomm.CommRequest) *urpccomm.CommResponse
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
type PipeLineFunc func(*urpccomm.CommRequest) error

// Register 绑定函数
func Register(reqBody interface{}, f interface{}) {
	defaultService.Register(reqBody, f)
}

// Group 返回指定名称的接口组
func Group(name string, pipes ...RequestPipeFunc) *InterfaceGroup {
	return defaultService.Group(name, pipes...)
}

// Run 启动默认服务
func Run(addr, errPrefix string) error {
	mcodePrefix = errPrefix
	return defaultService.Run(addr)
}

// DebugHttpHandler 返回GRPC调试处理
func DebugHttpHandler() http.Handler {
	return defaultService.DebugHttpHandler()
}

// UseHook 使用钩子列表,靠前的钩子最先进入,最后出来
func UseHook(hooks ...HookFunc) {
	defaultService.UseHook(hooks...)
}

// Stop 停止服务
func Stop() {
	defaultService.Stop()
}

// New  新建服务
func New() *Service {
	return newService()
}

// SetErrPrefix 设置错误码前缀
func SetErrPrefix(errPrefix string) {
	errPrefix = errPrefix
}

// DefaultService 返回默认的服务
func DefaultService() *Service {
	return defaultService
}
