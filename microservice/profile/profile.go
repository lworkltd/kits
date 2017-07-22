package profile

type Profile interface {
	Init()
}

type Base struct {
	GoMaxProcs float64 `toml:"go_max_procs"`
}

type Mongo struct {
	Url string `toml:"url"`
}

type Redis struct {
	Url      string `toml:"url"`
	Password string `toml:"password"`
	DB       int    `toml:"select_db"`
}

type Mysql struct {
	Url string `toml:"url"`
}

type Consul struct {
	Url string `toml:"url"`
}

type Service struct {
	Name       string `toml:"name"`
	Id         string `toml:"id"`
	Host       string `toml:"host"`
	PathPrefix string `toml:"path_prefix"`

	TraceEnabled     bool   `toml:"trace_enable"`
	McodeProfix      string `toml:"mcode_profix"`
	AccessLogEnabled bool   `toml:"access_log_enabled"`

	Reportable  bool     `toml:"reportable"`
	ReportIp    string   `toml:"report_ip"`
	ReportTags  []string `toml:"report_tags"`
	ReportHeach bool     `toml:"report_health"`

	Pprof_enabled     bool   `toml:"pprof_enabled"`
	Pprof_path_prefix string `toml:"pprof_path_prefix"`
}

type Discovery struct {
	EnableConsul   bool     `json:"enable_consul"`
	EnableStatic   bool     `json:"enable_static"`
	StaticServices []string `json:"static_service"`
}

type Svc struct {
	LoadBanlanceMode string `json:"load_balance_mode"`
	HystrixEnabled   bool   `json:"hytrix_enabled"`
	TracingEnabled   bool   `json:"traceing_enabled"`
	LoggerEnabled    bool   `json:"logger_enabled"`
}

type Hystrix struct {
	Url string `json:"url"`
}

type Zipkin struct {
	Url string `json:"url"`
}
