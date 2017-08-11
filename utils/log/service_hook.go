package log

import (
	"context"
	"log"
	"runtime"

	"github.com/Sirupsen/logrus"
	opentracing "github.com/opentracing/opentracing-go"
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
	entry.Data[ServiceTag] = hook.service
	entry.Data[ServiceIdTag] = hook.serverId
	entry.Data[EnvTag] = hook.env
	return nil
}

func (hook *ServiceTagHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// ServiceTagHook 是一个自定义的logrus.Hook实现，目的是在日志中展示日志所处的文件和行号
type FileLineHook struct {
	flag int
}

func NewFileLineHook(flag int) logrus.Hook {
	return &FileLineHook{
		flag: flag,
	}
}
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func makeFileLine(buf *[]byte, file string, line int, flag int) {
	if flag&(log.Lshortfile|log.Llongfile) != 0 {
		if flag&log.Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
	}
}

func (hook *FileLineHook) Fire(entry *logrus.Entry) error {
	fileineInfo := func() string {
		skip := 5
		if _, exist := entry.Data[DirectLoggerTag]; exist {
			skip = 6
			delete(entry.Data, DirectLoggerTag)
		}
		_, file, line, _ := runtime.Caller(skip)
		var buf []byte
		makeFileLine(&buf, file, line, hook.flag)
		return string(buf)
	}
	entry.Data[FilelineTag] = fileineInfo()

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
	entry.Data[TracingTag] = hook.tracingId

	return nil
}

func (hook *TracingTagHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

type TracingLogHook struct {
	tracingId string
}

func NewTracingLogHook() logrus.Hook {
	return &TracingLogHook{}
}

func (hook *TracingLogHook) Fire(entry *logrus.Entry) error {
	ctxValue, exist := entry.Data[ContextTag]
	if !exist {
		return nil
	}
	delete(entry.Data, ContextTag)
	ctx := ctxValue.(context.Context)
	sp := opentracing.SpanFromContext(ctx)
	if sp == nil {
		return nil
	}
	kvs := make([]interface{}, 0, len(entry.Data)*2)
	for fieldName, fieldValue := range entry.Data {
		kvs = append(kvs, fieldName)
		kvs = append(kvs, fieldValue)
	}

	if entry.Message != "" {
		kvs = append(kvs, "message")
		kvs = append(kvs, entry.Message)
	}

	if len(kvs) > 0 {
		sp.LogKV(kvs...)
	}

	return nil
}

func (hook *TracingLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
