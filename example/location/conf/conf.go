package conf

import (
	"strings"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/plugins"
	hystrixplugins "github.com/afex/hystrix-go/plugins"
	"github.com/lworkltd/kits/helper/consul"
	"github.com/lworkltd/kits/utils/eval"
	"github.com/lworkltd/kits/utils/ipnet"
	"github.com/lworkltd/kits/utils/jsonize"
	"github.com/lworkltd/kits/utils/log"
	"github.com/lworkltd/kits/service/discovery"
	"github.com/lworkltd/kits/service/invoke"
	"github.com/lworkltd/kits/service/profile"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

type Profile struct {
	Base     profile.Base
	Consul   profile.Consul
	Service  profile.Service
	Redis    profile.Redis
	Mongo    profile.Mongo
	Mysql    profile.Mysql
	Invoker  profile.Invoker
	Logger   profile.Logger
	Hystrix  profile.Hystrix
	Zipkin   profile.Zipkin
	Discover profile.Discovery
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
	var staticServices []*discovery.StaticService
	for _, line := range lines {
		words := strings.Split(line, " ")
		if len(words) == 0 {

		}
		service := &discovery.StaticService{
			Name:  words[0],
			Hosts: words[1:],
		}

		staticServices = append(staticServices, service)
	}
	s := discovery.NewStaticDiscovery(staticServices)

	return s.Discover, nil
}

func (pro *Profile) Init(tomlFile string) error {
	_, _, err := profile.Parse(tomlFile, pro)
	if err != nil {
		return err
	}

	if err := log.InitLoggerWithProfile(&pro.Logger); err != nil {
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
	var discoverOption discovery.Option
	discoverOption.RegisterFunc = consulClient.Register
	discoverOption.UnregisterFunc = consulClient.Unregister
	if pro.Discover.EnableConsul {
		discoverOption.SearchFunc = consulClient.Discover
		logrus.Debug("consul discovery enabled")
	}
	if pro.Discover.EnableStatic {
		staticsDiscover, err := makeStaticDiscover(pro.Discover.StaticServices)
		if err != nil {
			return err
		}
		discoverOption.StaticFunc = staticsDiscover
		logrus.Debug("static discovery enabled")
	}
	if err := discovery.Init(&discoverOption); err != nil {
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
		Discover:        discovery.Discover,
		LoadBalanceMode: pro.Invoker.LoadBanlanceMode,
		UseTracing:      pro.Invoker.TracingEnabled,
		UseCircuit:      pro.Invoker.CircuitEnabled,
		DoLogger:        pro.Invoker.LoggerEnabled,
		DefaultTimeout:  pro.Hystrix.DefaultTimeout,
		DefaultMaxConcurrentRequests: pro.Hystrix.DefaultMaxConcurrentRequests,
		DefaultErrorPercentThreshold: pro.Hystrix.DefaultErrorPercentThreshold,
	}
	if err := invoke.Init(invokeOption); err != nil {
		return err
	}

	return nil
}

func Dump() {
	mutiline := log.IsMultiLineFormat(configuration.Logger.Format)
	logrus.WithField("profile", jsonize.V(configuration, mutiline)).Info("Dump profile")
}

// 根据自己需要对方放出配置项目
// 返回服务配置
func GetService() *profile.Service {
	return &configuration.Service
}

// 获取redis配置
func GetRedis() *profile.Redis {
	return &configuration.Redis
}
