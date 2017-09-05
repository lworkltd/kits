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
	GoMaxProcs int8   `toml:"go_max_procs"` // 最大处理线程P的数量
	Mode       string `toml:"mode"`
}

// Mongo  用于初始化MongoDB的配置
type Mongo struct {
	Url string `toml:"url"`
}

// Redis 用于初始化Redis数据库的配置
//
type Redis struct {
	Endpoints string `toml:"endpoints"`
	Password  string `toml:"password"`
}

// Mysql 用于初始化Mysql数据库的配置
type Mysql struct {
	Url string `toml:"url"`
}

// Consul 用于初始化Consul的配置
// 如果你要使用Consul相关的eval操作，那么本配置是必须，否则在执行类似
// kv_of_consul等eval的时候将无法成功
type Consul struct {
	Url             string `toml:"endpoint"`
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
	Host        string `toml:"host"`         // 服务监听
	PathPrefix  string `toml:"path_prefix"`  // 接口前缀
	McodePrefix string `toml:"mcode_prefix"` // API错误码前缀

	TraceEnabled     bool `toml:"trace_enabled"`      // 启用OpenTrace,需要Zipkin配置
	AccessLogEnabled bool `toml:"access_log_enabled"` // 访问日志启用

	Reportable bool     `toml:"reportable"`  // 启用上报
	ReportIp   string   `toml:"report_ip"`   // 上报IP
	ReportTags []string `toml:"report_tags"` // 上报的标签
	ReportName string   `toml:"report_name"` // 上报名字
	ReportId   string   `toml:"report_id"`   // 上报的ID

	PprofEnabled    bool   `toml:"pprof_enabled"`     // 启用PPROF
	PprofPathPrefix string `toml:"pprof_path_prefix"` // PPROF的路径前缀,
	//srvContext log
	LogLevel        string `toml:"log_level"`
	LogFilePath     string `toml:"log_file_path"`
}

func (service *Service) BeforeParse() {
	service.TraceEnabled = true
	service.AccessLogEnabled = true
	service.Reportable = true
	service.ReportIp = "ip_of_first_interface()"
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
	EnableConsul   bool     `toml:"enable_consul"`   // 启用Consul，仅使用Consul时有效
	EnableStatic   bool     `toml:"enable_static"`   // 启用静态服务发现
	StaticServices []string `toml:"static_services"` // 静态服务配置,格式：["{serviceName} addr1 [addr2...]"]
}

func (discovery *Discovery) BeforeParse() {
	discovery.EnableConsul = true
	discovery.EnableStatic = true
	discovery.StaticServices = []string{}
}

func (discovery *Discovery) AfterParse() {
	if !discovery.EnableConsul && !discovery.EnableStatic {
		logrus.Warn("No discovery method enabled")
	}
}

// Invoker服务调用相关的配置
type Invoker struct {
	LoadBanlanceMode string `toml:"load_balance_mode"` // 负载均衡模式
	CircuitEnabled   bool   `toml:"circuit_enabled"`   // 启用Hystrix,需配置Hystrix才会生效
	TracingEnabled   bool   `toml:"traceing_enabled"`  // 启用Tracing，需配置Zipkin后有效
	LoggerEnabled    bool   `toml:"logger_enabled"`    // 启用日志打印，日志等级受控于
}

func (invoker *Invoker) BeforeParse() {
	invoker.LoadBanlanceMode = "round-robin"
	invoker.CircuitEnabled = true
	invoker.TracingEnabled = true
	invoker.LoggerEnabled = true
}
func (invoker *Invoker) AfterParse() {}

// Logger 日志配置
// 在几乎所有的服务或工具当中，这个配置项目都不应该缺席
type Logger struct {
	Format     string     `toml:"format"` // 日志的格式
	Level      string     `toml:"level"`
	File       string     `toml:"file"`
	TimeFormat string     `toml:"time_format"`
	LogFilePath    string `toml:"log_file_path"`
	Hooks      [][]string `toml:"hooks"`
}

func (logger *Logger) BeforeParse() {
	logger.Format = "json"
	logger.Level = "warn"
}
func (logger *Logger) AfterParse() {}

// Hystrix 熔断和异常请求的配置
type Hystrix struct {
	StatsdUrl             string `toml:"statsd_url"`
	Prefix                string `toml:"prefix"`
	Timeout               int    `toml:"timeout"`
	MaxConcurrentRequests int    `toml:"max_concurrent_request"`
	ErrorPercentThreshold int    `toml:"error_percent_threshold"`
}

func (hystrix *Hystrix) BeforeParse() {}
func (hystrix *Hystrix) AfterParse() {
	if hystrix.StatsdUrl == "" {
		logrus.WithFields(logrus.Fields{
			"profile": "Hystrix",
		}).Warn("StatsdUrl is still empty")
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
