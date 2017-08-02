package log

import (
	"fmt"
	"runtime"

	"github.com/Sirupsen/logrus"
)

// ServiceTagHook 是一个自定义的logrus.Hook实现，目的是为日志添加服务名称，服务ID和环境名称
type ServiceTagHook struct {
	service  string
	serverId string
	env      string
}

func NewServiceTagHook(service, serviceId, env string) logrus.Hook {
	return &ServiceTagHook{
		service:  service,
		serverId: serviceId,
		env:      env,
	}
}

func (hook *ServiceTagHook) Fire(entry *logrus.Entry) error {
	entry.Data["@service"] = hook.service
	entry.Data["@serviceid"] = hook.serverId
	entry.Data["@environment"] = hook.serverId
	return nil
}

func (hook *ServiceTagHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// ServiceTagHook 是一个自定义的logrus.Hook实现，目的是在日志中展示日志所处的文件和行号
type FileLineHook struct {
	short bool
}

func NewFileLineHook(short bool) logrus.Hook {
	return &FileLineHook{
		short: short,
	}
}

var DirectLoggerTag = "@directLoggerTag"

func (hook *FileLineHook) Fire(entry *logrus.Entry) error {
	fileineInfo := func() string {
		skip := 5
		if _, exist := entry.Data[DirectLoggerTag]; exist {
			skip = 6
			delete(entry.Data, DirectLoggerTag)
		}
		_, file, line, _ := runtime.Caller(skip)
		// TODO:short file path
		return fmt.Sprintf("%s:%d", file, line)
	}
	entry.Data["@fileline"] = fileineInfo()

	return nil
}

func (hook *FileLineHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

type TracingTagHook struct {
	tracingId string
}

func NewTracingTagHook(id string) logrus.Hook {
	return &TracingTagHook{
		tracingId: id,
	}
}

func (hook *TracingTagHook) Fire(entry *logrus.Entry) error {
	entry.Data["@tracing"] = hook.tracingId

	return nil
}

func (hook *TracingTagHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
