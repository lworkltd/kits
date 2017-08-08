package context

import (
	"net/http"

	"github.com/lworkltd/kits/utils/log"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
)

const (
	TraceIdHeader = "X-B3-Traceid"
	SpanIdHeader  = "X-B3-Traceid"
)

type Context interface {
	context.Context
	logrus.FieldLogger
	Tracer

	Inject(header http.Header)
	SubContext(string) Context
}

func FromHttpRequest(request *http.Request, logger logrus.FieldLogger) Context {
	var sp opentracing.Span
	name := "http://xx/xx"
	// TODO: fix the name of tracing span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(request.Header))
	if err != nil {
		sp = opentracing.StartSpan(name)
	} else {
		sp = opentracing.StartSpan(name, ext.RPCServerOption(wireContext))
	}

	return &tracingCtx{
		Context:     opentracing.ContextWithSpan(context.Background(), sp),
		FieldLogger: logger,
	}
}

var (
	_ Context            = new(tracingCtx)
	_ logrus.FieldLogger = new(tracingCtx)
	_ Tracer             = new(tracingCtx)
)

type Tracer interface {
	Finish()
	TracingId() string
	SpanId() string
}

type tracingCtx struct {
	context.Context
	logrus.FieldLogger
	tracingId string
	spanId    string
}

// Finish 结束当前Tracing Span
func (ctx *tracingCtx) Finish() {
	opentracing.SpanFromContext(ctx.Context).Finish()
}

// Inject 将tracing信息注入Http的头部，用于网间传输
func (ctx *tracingCtx) Inject(header http.Header) {
	opentracing.GlobalTracer().Inject(
		opentracing.SpanFromContext(ctx.Context).Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header))
}

// SubContext 启用一个子Span
func (ctx *tracingCtx) SubContext(name string) Context {
	sp := opentracing.StartSpan(
		name,
		opentracing.ChildOf(
			opentracing.SpanFromContext(ctx.Context).Context(),
		),
	)
	return &tracingCtx{
		Context:     opentracing.ContextWithSpan(ctx.Context, sp),
		FieldLogger: ctx.FieldLogger,
	}
}

// TracingId 获取当前的Tracing
func (ctx *tracingCtx) TracingId() string {
	if ctx.tracingId == "" {
		headers := http.Header{}
		ctx.Inject(headers)
		ctx.tracingId = headers[TraceIdHeader][0]
	}

	return ctx.tracingId
}

// SpanId 获取当前的Span Id
func (ctx *tracingCtx) SpanId() string {
	if ctx.spanId == "" {
		headers := http.Header{}
		ctx.Inject(headers)
		ctx.spanId = headers[SpanIdHeader][0]
	}

	return ctx.spanId
}

var directLogFields = logrus.Fields{
	log.DirectLoggerTag: true,
}

func contextLogFields(ctx context.Context) logrus.Fields {
	return logrus.Fields{
		log.ContextTag: ctx,
	}
}
func (ctx *tracingCtx) WithField(key string, value interface{}) *logrus.Entry {
	return ctx.FieldLogger.WithFields(contextLogFields(ctx)).WithField(key, value)
}

func (ctx *tracingCtx) WithFields(fields logrus.Fields) *logrus.Entry {
	return ctx.FieldLogger.WithFields(contextLogFields(ctx)).WithFields(fields)
}

func (ctx *tracingCtx) WithError(err error) *logrus.Entry {
	return ctx.FieldLogger.WithFields(contextLogFields(ctx)).WithError(err)
}

func (ctx *tracingCtx) Debugf(format string, args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Debugf(format, args...)
}

func (ctx *tracingCtx) Infof(format string, args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Infof(format, args...)
}

func (ctx *tracingCtx) Printf(format string, args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Printf(format, args...)
}

func (ctx *tracingCtx) Warnf(format string, args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Warnf(format, args...)
}

func (ctx *tracingCtx) Warningf(format string, args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Warningf(format, args...)
}

func (ctx *tracingCtx) Errorf(format string, args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Errorf(format, args...)
}

func (ctx *tracingCtx) Fatalf(format string, args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Fatalf(format, args...)
}

func (ctx *tracingCtx) Panicf(format string, args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Panicf(format, args...)
}

func (ctx *tracingCtx) Debug(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Debug(args...)
}

func (ctx *tracingCtx) Info(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Info(args...)
}

func (ctx *tracingCtx) Print(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Print(args...)
}

func (ctx *tracingCtx) Warn(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Warn(args...)
}

func (ctx *tracingCtx) Warning(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Warning(args...)
}

func (ctx *tracingCtx) Error(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Error(args...)
}

func (ctx *tracingCtx) Panic(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Panic(args...)
}

func (ctx *tracingCtx) Debugln(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Debugln(args...)
}
func (ctx *tracingCtx) Infoln(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Infoln(args...)
}

func (ctx *tracingCtx) Println(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Println(args...)
}

func (ctx *tracingCtx) Warnln(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Warnln(args...)
}

func (ctx *tracingCtx) Warningln(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Warningln(args...)
}

func (ctx *tracingCtx) Errorln(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Errorln(args...)
}

func (ctx *tracingCtx) Fatalln(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Fatalln(args...)
}

func (ctx *tracingCtx) Panicln(args ...interface{}) {
	ctx.FieldLogger.WithFields(directLogFields).WithFields(contextLogFields(ctx)).Panicln(args...)
}
