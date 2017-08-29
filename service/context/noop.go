package context

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
)

type NoopContext struct {
	context.Context
	logrus.FieldLogger
	Tracer
}

type NoopTracer struct {
}

func (tracer *NoopTracer) Finish()            {}
func (tracer *NoopTracer) Inject(http.Header) {}
func (tracer *NoopTracer) TracingId() string  { return "" }
func (tracer *NoopTracer) SpanId() string     { return "" }
func (ctx *NoopContext) SubContext(string) Context {
	return &NoopContext{
		FieldLogger: &NoopLogger{},
		Context:     context.Background(),
		Tracer:      &NoopTracer{},
	}
}
func (ctx *NoopContext) Inject(http.Header) {}

type NoopLogger struct {
}

func (logger *NoopLogger) WithField(key string, value interface{}) *logrus.Entry {
	defer recover()
	return &logrus.Entry{
		Logger: &logrus.Logger{},
	}
}

func (logger *NoopLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	defer recover()
	return &logrus.Entry{
		Logger: &logrus.Logger{},
	}
}
func (logger *NoopLogger) WithError(err error) *logrus.Entry {
	defer recover()
	return &logrus.Entry{
		Logger: &logrus.Logger{},
	}
}

func (logger *NoopLogger) Debugf(format string, args ...interface{})   {}
func (logger *NoopLogger) Infof(format string, args ...interface{})    {}
func (logger *NoopLogger) Printf(format string, args ...interface{})   {}
func (logger *NoopLogger) Warnf(format string, args ...interface{})    {}
func (logger *NoopLogger) Warningf(format string, args ...interface{}) {}
func (logger *NoopLogger) Errorf(format string, args ...interface{})   {}
func (logger *NoopLogger) Fatalf(format string, args ...interface{})   {}
func (logger *NoopLogger) Panicf(format string, args ...interface{})   {}

func (logger *NoopLogger) Debug(args ...interface{})   {}
func (logger *NoopLogger) Info(args ...interface{})    {}
func (logger *NoopLogger) Print(args ...interface{})   {}
func (logger *NoopLogger) Warn(args ...interface{})    {}
func (logger *NoopLogger) Warning(args ...interface{}) {}
func (logger *NoopLogger) Error(args ...interface{})   {}
func (logger *NoopLogger) Fatal(args ...interface{})   {}
func (logger *NoopLogger) Panic(args ...interface{})   {}

func (logger *NoopLogger) Debugln(args ...interface{})   {}
func (logger *NoopLogger) Infoln(args ...interface{})    {}
func (logger *NoopLogger) Println(args ...interface{})   {}
func (logger *NoopLogger) Warnln(args ...interface{})    {}
func (logger *NoopLogger) Warningln(args ...interface{}) {}
func (logger *NoopLogger) Errorln(args ...interface{})   {}
func (logger *NoopLogger) Fatalln(args ...interface{})   {}
func (logger *NoopLogger) Panicln(args ...interface{})   {}

var _ logrus.FieldLogger = new(NoopLogger)
var _ Context = new(NoopContext)
var _ Tracer = new(NoopTracer)
