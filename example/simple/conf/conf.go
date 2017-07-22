package conf

type Application struct {
	MongoUrl string `env:"mongo.url" toml:"mongo.url"`
}

type Mongo struct {
}

type Redis struct {
}

type Zipkin struct {
}

type Discovery struct {
}

type Consul struct {
}
