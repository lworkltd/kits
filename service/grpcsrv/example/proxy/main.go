package main

import (
	"context"
	"fmt"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/grpcinvoke"
	"github.com/lworkltd/kits/service/grpcsrv"
	"github.com/lworkltd/kits/service/grpcsrv/example/testproto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
)

// ProxyCheck 代理检查
func ProxyCheck(ctx context.Context, commReq *grpccomm.CommRequest) error {
	return nil
}

// DynamicProxy 动态代理函数
func DynamicProxy(ctx context.Context, req *grpccomm.CommRequest) (grpcinvoke.Service, error) {
	header := &grpccomm.CommHeader{}
	if header.BaseInfo == nil {
		return nil, code.NewMcode("TENANT_NOT_FOUND", "tenant not found")
	}

	addr := fmt.Sprintln("gateway.MT4-%s", strings.ToUpper(header.BaseInfo.TenantId))

	return grpcinvoke.Addr(addr), nil
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})
	// 静态代理
	staticProxyGroup := grpcsrv.Group("static-proxy", ProxyCheck)
	// 自定义请求匹配规则，匹配则转发到指定微服务
	staticProxyGroup.ProxyService("Foo", grpcinvoke.Name("FooServiceV3"), func(reqName string) bool {
		return strings.HasPrefix(reqName, "FooV3")
	})

	// 将指定接口名称，将请求归属于列表中的请求转发到指定微服务
	staticProxyGroup.ProxyService("Foo", grpcinvoke.Name("FooServiceV2"),
		&testproto.ReportRequest{},
		"DeleteUserRequest",
	)

	// 一律全部转发到指定微服务
	staticProxyGroup.ProxyService("Foo", grpcinvoke.Name("FooServiceV1"))

	// 指定接口，不考虑请求服务，将在接口列表中的请求类型转发到指定微服务
	staticProxyGroup.ProxyInterface(grpcinvoke.Name("BarService"), &testproto.DepositRequest{})
	// 指定接口，，不考虑请求服务，转发到指定地址服务
	staticProxyGroup.ProxyInterface(grpcinvoke.Addr("www.baidu.com"), "SearchRequest")

	// 动态代理
	// 指定
	dynamicProxyGroup := grpcsrv.Group("dynamic-proxy", ProxyCheck)
	// 转发请求服务为Bridge到实时计算的目标微服务
	dynamicProxyGroup.ProxyService("Bridge", DynamicProxy, "DepositRequest")
	// 转发接口到实时计算出的目标微服务
	dynamicProxyGroup.ProxyInterface(grpcsrv.DynamicProxyTarget(DynamicProxy), "SearchRequest")
	// 自定义代理，根据请求内容返回是否匹配此代理规则，若匹配则返回服务
	dynamicProxyGroup.Proxy(func(ctx context.Context, req *grpccomm.CommRequest) (bool, grpcinvoke.Service, error) {
		if req.ReqSercice == "Bridge" && req.ReqInterface == "DepositRequest" {
			return false, grpcinvoke.Name("mt4-bridge"), nil
		}
		return false, nil, nil
	})

	// For pprof
	go func() {
		http.Handle("/grpcsrv", grpcsrv.DebugHttpHandler())
		http.ListenAndServe(":8080", nil)
	}()

	// 监听
	grpcsrv.Run("0.0.0.0:8090", "TESTECHO_")
}
