package profile

type Profile interface {
	Init()
}

type BaseProfile struct {
	GoMaxProcs float64 `toml:"go_max_procs"`
}

type MongoProfile struct {
	Url string `toml:"url"`
}

type RedisProfile struct {
	Url      string `toml:"url"`
	Password string `toml:"password"`
	DB       int    `toml:"select_db"`
}

type MysqlProfile struct {
	Url string `toml:"url"`
}

type ConsulProfile struct {
	Url string `toml:"url"`
}

type ServiceProfile struct {
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
