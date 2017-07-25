package conf

type Application struct {
	// 添加你的配置
}

func GetApplication() *Application {
	return &configuration.Application
}
