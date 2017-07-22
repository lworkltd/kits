package conf

type Profile struct {
	MongoUrl string `env:"mongo.url" toml:"mongo.url"`
}
