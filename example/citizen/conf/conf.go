package conf

import (
	"strings"

	"fmt"
	"github.com/afex/hystrix-go/plugins"
	hystrixplugins "github.com/afex/hystrix-go/plugins"
	"github.com/lvhuat/kits/helper/consul"
	"github.com/lvhuat/kits/pkgs/eval"
	"github.com/lvhuat/kits/pkgs/ipnet"
	"github.com/lvhuat/kits/service/discover"
	"github.com/lvhuat/kits/service/invoke"
	"github.com/lvhuat/kits/service/profile"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

type Profile struct {
	Base        profile.Base
	Consul      profile.Consul
	Service     profile.Service
	Redis       profile.Redis
	Mongo       profile.Mongo
	Mysql       profile.Mysql
	Invoker     profile.Invoker
	Logger      profile.Logger
	Hystrix     profile.Hystrix
	Zipkin      profile.Zipkin
	Discover    profile.Discovery
	Application Application
}

var configuration Profile

func Parse(f ...string) error {
	fileName := "app.toml"
	if len(f) > 0 {
		fileName = f[0]
	}
	return configuration.Init(fileName)
}

func makeStaticDiscover(lines []string) (func(string) ([]string, []string, error), error) {
	var staticServices []*discover.StaticService
	for _, line := range lines {
		words := strings.Split(line, " ")
		if len(words) == 0 {

		}
		service := &discover.StaticService{
			Name:  words[0],
			Hosts: words[1:],
		}

		staticServices = append(staticServices, service)
	}
	s := discover.NewStaticDiscoverer(staticServices)

	return s.Discover, nil
}

func (pro *Profile) Init(tomlFile string) error {
	_, _, err := profile.Parse(tomlFile, pro)
	if err != nil {
		return err
	}

	consulClient, err := consul.New(pro.Consul.Url)
	if err != nil {
		return err
	}
	consul.SetClient(consulClient)

	// 注册eval解析器
	eval.RegisterKeyValueExecutor("kv_of_consul", consulClient.KeyValue)
	eval.RegisterKeyValueExecutor("ip_of_interface", ipnet.Ipv4)

	// 将填充配置中使用了eval语法
	if err := eval.Complete(&pro); err != nil {
		return err
	}

	// Discover 服务发现
	var discoverOption discover.Option
	if pro.Discover.EnableConsul {
		discoverOption.SearchFunc = consulClient.Discover
	}
	if pro.Discover.EnableStatic {
		staticsDiscover, err := makeStaticDiscover(pro.Discover.StaticServices)
		if err != nil {
			return err
		}
		discoverOption.StaticFunc = staticsDiscover
	}
	if err := discover.Init(&discoverOption); err != nil {
		return err
	}

	// Hystrix 熔断初始化
	if pro.Hystrix.StatsdUrl != "" {
		hystrixplugins.InitializeStatsdCollector(&plugins.StatsdCollectorConfig{
			StatsdAddr: pro.Hystrix.StatsdUrl,
			Prefix:     pro.Hystrix.Prefix,
		})
	}

	// Zipkin 调用追踪
	if pro.Zipkin.Url != "" {
		collector, err := zipkin.NewHTTPCollector(pro.Zipkin.Url)
		if err != nil {
			return err
		}

		// Create our recorder.
		recorder := zipkin.NewRecorder(
			collector,
			pro.Base.Mode == "debug",
			pro.Service.Host,
			fmt.Sprintf("%s:%s", pro.Service.ReportId, pro.Service.ReportName),
		)

		// Create our tracer.
		tracer, err := zipkin.NewTracer(
			recorder,
			// zipkin.ClientServerSameSpan(sameSpan),
			// zipkin.TraceID128Bit(traceID128Bit),
		)
		if err != nil {
			return err
		}

		// Explicitly set our tracer to be the default tracer.
		opentracing.InitGlobalTracer(tracer)
	}

	// Invoker 服务调用初始化
	invokeOption := &invoke.Option{
		Discover:        discover.Discover,
		LoadBalanceMode: pro.Invoker.LoadBanlanceMode,
		UseTracing:      pro.Invoker.TracingEnabled,
		UseCircuit:      pro.Invoker.CircuitEnabled,
		DoLogger:        pro.Invoker.LoggerEnabled,
	}
	if err := invoke.Init(invokeOption); err != nil {
		return err
	}

	return nil
}

// 根据自己需要对方放出配置项目

func GetService() *profile.Service {
	return &configuration.Service
}
