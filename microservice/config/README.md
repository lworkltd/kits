config
-----
配置文件读取，支持环境变量+toml
环境变量优先


例如
----
```
type Profile struct {
	MongoUrl            string `env:"MONGO_URL" toml:"MONGO_URL"`
	WhiteList           []string `env:"WHITE_LIST" toml:"WHITE_LIST"`
	DiscoveryHost       string `env:"DISCOVER_HOST" toml:"DISCOVER.HOST"`
	DiscoveryPort       int `env:"DISCOVER_PORT" toml:"DISCOVER.PORT"`
	DiscoveryTags       []string`env:"DISCOVER_TAGS" toml:"DISCOVER.TAGS"`
}
```



----
待完善
1.toml子级尚未支持