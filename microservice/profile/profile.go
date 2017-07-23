package profile

import (
	"github.com/Sirupsen/logrus"
)

// Profile 用于适配配置项目的接口，对于实现了此接口的配置项目，将在配置加载后执行Init
// 这样一来，适配的对象将有机会在没有充分的配置时使用默认配置
type Profile interface {
	BeforeParse() // 初始化配置的值
	AfterParse()  // 检查默认值
}

// Base 是进程在运行时的配置
type Base struct {
	GoMaxProcs float64 `toml:"go_max_procs"` // 最大处理线程P的数量
}

// Mongo  用于初始化MongoDB的配置
type Mongo struct {
	Url string `toml:"url"`
}

// Redis 用于初始化Redis数据库的配置
//
type Redis struct {
	Url      string `toml:"url"`
	Password string `toml:"password"`
	DB       int    `toml:"select_db"`
}

// Mysql 用于初始化Mysql数据库的配置
type Mysql struct {
	Url string `toml:"url"`
}

// Consul 用于初始化Consul的配置
// 如果你要使用Consul相关的eval操作，那么本配置是必须，否则在执行类似
// kv_of_consul等eval的时候将无法成功
type Consul struct {
	Url             string `toml:"url"`
	AutoSyncEnabled bool   `toml:"auto_sync_enabled"`
}

func (consul *Consul) BeforeParse() {
	consul.AutoSyncEnabled = true
}

func (consul *Consul) AfterParse() {
	if consul.Url == "" {
		logrus.WithFields(logrus.Fields{
			"profile": "Consul",
		}).Error("Url is empty")
	}
}

// Service 用于初始化服务的配置
// 如果
type Service struct {
	Host        string `toml:"host"`         // 服务前缀
	PathPrefix  string `toml:"path_prefix"`  // 接口前缀
	McodeProfix string `toml:"mcode_profix"` // API错误码前缀

	TraceEnabled     bool `toml:"trace_enable"`       // 启用OpenTrace,需要Zipkin配置
	AccessLogEnabled bool `toml:"access_log_enabled"` // 访问日志启用

	Reportable  bool     `toml:"reportable"`    // 启用上报
	ReportIp    string   `toml:"report_ip"`     // 上报IP
	ReportTags  []string `toml:"report_tags"`   // 上报的标签
	ReportHeach bool     `toml:"report_health"` // 启用健康检查
	ReportName  string   `toml:"report_name"`   // 上报名字
	ReportId    string   `toml:"report_id"`     // 上报的ID

	PprofEnabled    bool   `toml:"pprof_enabled"`     // 启用PPROF
	PprofPathPrefix string `toml:"pprof_path_prefix"` // PPROF的路径前缀,
}

func (service *Service) BeforeParse() {
	service.TraceEnabled = true
	service.AccessLogEnabled = true
	service.Reportable = true
	service.ReportIp = "ip_of_first_interface()"
	service.ReportHeach = true
	service.PprofEnabled = true
}

func (service *Service) AfterParse() {
	if service.Reportable {
		if service.ReportIp == "" || service.ReportName == "" || service.ReportId == "" {
			logrus.WithFields(logrus.Fields{
				"profile": "Service",
			}).Error("Reportable is set,you need a report ip, name and id")
		}
	}
}

type Discovery struct {
	EnableConsul   bool     `json:"enable_consul"`  // 启用Consul，仅使用Consul时有效
	EnableStatic   bool     `json:"enable_static"`  // 启用静态服务发现
	StaticServices []string `json:"static_service"` // 静态服务配置,格式：["{serviceName} addr1 [addr2...]"]
}

// Svc服务调用相关的配置
type Svc struct {
	LoadBanlanceMode string `json:"load_balance_mode"` // 负载均衡模式
	HystrixEnabled   bool   `json:"hytrix_enabled"`    // 启用Hystrix,需配置Hystrix才会生效
	TracingEnabled   bool   `json:"traceing_enabled"`  // 启用Tracing，需配置Zipkin后有效
	LoggerEnabled    bool   `json:"logger_enabled"`    // 启用日志打印，日志等级受控于
}

func (svc *Svc) BeforeParse() {
	svc.LoadBanlanceMode = "round-robin"
	svc.HystrixEnabled = true
	svc.TracingEnabled = true
	svc.LoggerEnabled = true
}
func (svc *Svc) AfterParse() {}

// Logger 日志配置
// 在几乎所有的服务或工具当中，这个配置项目都不应该缺席
type Logger struct {
	Format string `json:"format"` // 日志的格式
	Level  string `json:"level"`
}

func (logger *Logger) BeforeParse() {
	logger.Format = "json"
	logger.Level = "warn"
}
func (logger *Logger) AfterParse() {}

// Hystrix 熔断和异常请求的配置
type Hystrix struct {
	Url                   string `toml:"url"`
	Timeout               int    `toml:"timeout"`
	MaxConcurrentRequests int    `toml:"max_concurrent_request"`
	ErrorPercentThreshold int    `toml:"error_percent_threshold"`
}

func (hystrix *Hystrix) BeforeParse() {}
func (hystrix *Hystrix) AfterParse() {
	if hystrix.Url == "" {
		logrus.WithFields(logrus.Fields{
			"profile": "Hystrix",
		}).Warn("Url is still empty")
	}
}

// Zipkin OpenTracing 的配置
type Zipkin struct {
	Url string `json:"url"` // Zipkin服务的地址
}

func (zipkin *Zipkin) BeforeParse() {}
func (zipkin *Zipkin) AfterParse() {
	if zipkin.Url == "" {
		logrus.WithFields(logrus.Fields{
			"profile": "Zipkin",
		}).Warn("Url is empty")
	}
}
