config
-----
配置文件读取，支持环境变量+toml
环境变量优先


例如
----


使用方法
----
1.准备好你的app.toml
2.准备你的Profile结构
```
type Profile struct {
	Base    profile.Base
	Service profile.Service
	Redis   profile.Redis
	Mongo   profile.Mongo
	Mysql   profile.Mysql
	Consul  profile.Consul
}
```

3.profile.Parse("app.toml")

