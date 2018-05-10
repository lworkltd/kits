package urpcsrv

import (
	"reflect"

	context "golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/grpcinvoke"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/lworkltd/kits/service/urpcsrv/urpccomm"
)

// RequestPipeFunc 通用预处理
type RequestPipeFunc func(context.Context, *urpccomm.CommRequest) error

// InterfaceGroup 通用分组
type InterfaceGroup struct {
	name      string
	pipelines []RequestPipeFunc
	infos     map[string]*RegisterInfo
	parent    *InterfaceGroup
	service   *Service

	level   int
	parents []*InterfaceGroup
}

// Use 添加预处理
func (group *InterfaceGroup) Use(pipes ...RequestPipeFunc) {
	group.pipelines = append(group.pipelines, pipes...)
}

func (group *InterfaceGroup) doPipe(ctx context.Context, commReq *urpccomm.CommRequest) error {
	for _, g := range group.parents {
		for _, p := range g.pipelines {
			err := p(ctx, commReq)
			if err != nil {
				return err
			}
		}
	}

	for _, p := range group.pipelines {
		err := p(ctx, commReq)
		if err != nil {
			return err
		}
	}

	return nil
}

// RpcRequest 请求处理
func (group *InterfaceGroup) RpcRequest(ctx context.Context, commReq *urpccomm.CommRequest) *urpccomm.CommResponse {
	// 预处理
	if err := group.doPipe(ctx, commReq); err != nil {
		return newRspFromError(err)
	}

	// 查找对应函数
	regInfo, exist := group.infos[commReq.ReqInterface]
	if !exist {
		return newErrorRsp("GRPC_METHOD_NOTFOUND", "grpc method %s not found", commReq.ReqInterface)
	}

	var (
		bodyMsg   reflect.Value
		headerMsg = emptyHeader
	)

	// 解析头部
	if regInfo.newHeader != nil {
		headerMsg = regInfo.newHeader()
	}
	if len(commReq.Header) > 0 && regInfo.newHeader != nil {
		err := proto.Unmarshal(commReq.Header, headerMsg.Interface().(proto.Message))
		if err != nil {
			return newErrorRsp("GRPC_METHOD_BADHeader", "grpc method %s parse header failed", commReq.ReqInterface)
		}
	}

	// 解析内容
	if regInfo.newBody != nil {
		bodyMsg = regInfo.newBody()
		if len(commReq.Body) > 0 {
			err := proto.Unmarshal(commReq.Body, bodyMsg.Interface().(proto.Message))
			if err != nil {
				return newErrorRsp("GRPC_METHOD_BADBODY", "grpc method %s parse body failed", commReq.ReqInterface)
			}
		}
	}

	return call0(ctx, headerMsg, bodyMsg, commReq, regInfo)
}

// Register 注册
func (group *InterfaceGroup) Register(reqBody interface{}, f interface{}) {
	regInfo := createRegInfo(reqBody, f)
	group.service.regInterface(group, regInfo.reqName)
	debugRegisterInfo(regInfo)
	group.infos[regInfo.reqName] = regInfo
}

// ProxyTarget 代理目标
type ProxyTarget struct {
	detect ProxyDetect
}

type serviceProxyMatcher struct {
	serviceName string
	match       func(string) bool
	target      DynamicProxyTarget
}

func (serviceProxyMatcher *serviceProxyMatcher) Match(ctx context.Context, req *urpccomm.CommRequest) (bool, grpcinvoke.Service, error) {
	if serviceProxyMatcher.serviceName != "" && serviceProxyMatcher.serviceName != req.ReqSercice {
		return false, nil, nil
	}

	if serviceProxyMatcher.match != nil {
		if matched := serviceProxyMatcher.match(req.ReqInterface); !matched {
			return false, nil, nil
		}
	}

	service, err := serviceProxyMatcher.target(ctx, req)
	if err != nil {
		return true, nil, code.NewMcodef("GRPC_PROXY_ERROR", "proxy target:%v", err)
	}

	return true, service, nil
}

func reqNameFrom(req interface{}) string {
	s, ok := req.(string)
	if ok {
		return s
	}

	rt := reflect.TypeOf(req)
	if rt.Kind() != reflect.Ptr {
		unexpectError("req type must be string or pointer,got %s", rt.String())
	}

	return rt.Elem().String()
}

func newServiceProxyDetecter(serviceName string, target interface{}, reqs ...interface{}) *serviceProxyMatcher {
	matcher := &serviceProxyMatcher{
		serviceName: serviceName,
	}

	if len(reqs) > 0 {
		// 匹配函数
		match, ok := reqs[0].(func(string) bool)
		if ok {
			matcher.match = match
		} else {
			// 列表
			reqNames := make(map[string]bool, len(reqs))
			for _, req := range reqs {
				reqNames[reqNameFrom(req)] = true
			}

			matcher.match = func(reqName string) bool {
				return reqNames[reqName]
			}
		}

	}

	matcher.target = mustParseTarget(target)

	return matcher
}

// DynamicProxyTarget 动态代理目标获取
type DynamicProxyTarget func(ctx context.Context, req *urpccomm.CommRequest) (grpcinvoke.Service, error)

func mustParseTarget(target interface{}) DynamicProxyTarget {
	targetFinder, ok := target.(func(ctx context.Context, req *urpccomm.CommRequest) (grpcinvoke.Service, error))
	if ok {
		return targetFinder
	}

	targetFinder, ok = target.(DynamicProxyTarget)
	if ok {
		return targetFinder
	}

	service, ok := target.(grpcinvoke.Service)
	if ok {
		return func(ctx context.Context, req *urpccomm.CommRequest) (grpcinvoke.Service, error) {
			return service, nil
		}
	}

	unexpectError(`proxy target not support,candidates is:
		- grpcinvoke.Service
		- func(context.Context, *urpccomm.CommRequest) (grpcinvoke.Service, error)
		got %v`, reflect.TypeOf(target).String())

	return nil
}

type proxyMatcher struct {
}

// ProxyDetect 代理判决
type ProxyDetect func(context.Context, *urpccomm.CommRequest) (bool, grpcinvoke.Service, error)

// Name 返回名称
func (group *InterfaceGroup) Name() string {
	return group.name
}

// Group 获取group
func (group *InterfaceGroup) Group(name string, pipes ...RequestPipeFunc) *InterfaceGroup {
	subGroup, isNew := group.service.newGroup(name, pipes...)
	if isNew {
		subGroup.parent = group
		subGroup.level = group.level + 1
		subGroup.parents = append(group.parents, group)
	}

	return subGroup
}
