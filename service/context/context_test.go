package context

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"testing"
	"time"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/invoke"
	"github.com/lworkltd/kits/service/restful/code"
	logutils "github.com/lworkltd/kits/utils/log"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

func TestContextDebug(t *testing.T) {
	collector, err := zipkin.NewHTTPCollector("http://47.90.65.243:9411/api/v1/spans")
	if err != nil {
		fmt.Printf("unable to create Zipkin HTTP collector: %+v\n", err)
		os.Exit(-1)
	}

	// Create our recorder.
	recorder := zipkin.NewRecorder(collector, false, "127.0.0.1:8083", "qwerty")

	// Create our tracer.
	tracer, err := zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(true),
		zipkin.TraceID128Bit(true),
	)
	opentracing.InitGlobalTracer(tracer)

	if err != nil {
		fmt.Printf("unable to create Zipkin tracer: %+v\n", err)
		os.Exit(-1)
	}

	pullDataFromServiceB := func(ctx Context, dbData interface{}, parameters ...interface{}) (interface{}, code.Error) {
		url := "111/any/path"
		service := "service-a"
		serviceId := "service-a-1"
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			ctx.WithError(err).Info("pullDataFromServiceB failed")
			return nil, code.Newf(1234, "bad url %v", err)
		}

		subName := fmt.Sprintf("http://%s-%s:%s", service, serviceId, url)
		subContext := ctx.SubContext(subName)
		defer subContext.Finish()

		ctx.Inject(request.Header)
		client := http.Client{}
		_, err = client.Do(request)
		if err != nil {
			subContext.WithError(err).Info("pullDataFromServiceB failed")
			return nil, code.Newf(1234, "bad request %v", err)
		}

		// parse data from response ...

		ctx.Info("pullDataFromServiceB done")
		return map[string]interface{}{
			"anyfield": 1,
		}, nil
	}

	readDataFromDatabase := func(ctx Context, parameters ...interface{}) (interface{}, code.Error) {
		dbContext := ctx.SubContext("read-data-base")
		defer dbContext.Finish()

		// Read data from database ...
		dbContext.Info("readDataFromDatabase done")
		return map[string]interface{}{
			"anyfield": 1,
		}, nil
	}

	readData := func(ctx Context, parameters ...interface{}) (interface{}, code.Error) {
		data, cerr := readDataFromDatabase(ctx, parameters...)
		if cerr != nil {
			ctx.WithFields(logrus.Fields{
				"parameters": parameters,
				"error":      cerr,
			}).Error("Read data base failed")
			return nil, cerr
		}
		ctx.WithError(fmt.Errorf("xxx")).WithField("abc", "123").Info("readDataFromDatabase OK")
		ctx.WithFields(logrus.Fields{"abc": "123"}).Info("readDataFromDatabase OK")

		logrus.Error("##################")
		return pullDataFromServiceB(ctx, data, parameters...)
	}

	handler := func(serviceCtx Context, r *gin.Context) (interface{}, code.Error) {
		if r.Query("panic") != "" {
			serviceCtx.Error("Panic in handler")
			panic("panic test")
		}

		if r.Query("error") != "" {
			serviceCtx.Error("Error in handler")
			return nil, code.New(1002, "error test")
		}

		return readData(serviceCtx, r.Query("name"), r.Query("type"))
	}

	wrapFunc := func(f func(Context, *gin.Context) (interface{}, code.Error)) func(*gin.Context) {
		return func(httpCtx *gin.Context) {
			Prefix := "SERVICE_A" // 错误码前缀
			logger := logrus.New()
			// 设置日志等级
			logger.Level = logrus.InfoLevel
			// 设置日志格式,让附加的TAG放在最前面
			logger.Formatter = &logutils.TextFormatter{
				TimestampFormat: "01-02 15:04:05.999",
			}

			// 附加服务ID
			logger.Hooks.Add(logutils.NewServiceTagHook("service-a", "service-a-10", "dev"))
			// 附加日志文件行号
			logger.Hooks.Add(logutils.NewFileLineHook(log.Lshortfile))
			// 附加Tracing TAG
			logger.Hooks.Add(logutils.NewTracingLogHook())
			serviceCtx := FromHttpRequest(httpCtx.Request, logger)
			defer serviceCtx.Finish()

			// 附加Tracing Id
			tracingHeader := http.Header{}
			serviceCtx.Inject(tracingHeader)
			logger.Hooks.Add(logutils.NewTracingTagHook(tracingHeader[traceIdHeader][0]))

			since := time.Now()
			var (
				data interface{}
				cerr code.Error
			)
			defer func() {
				// 拦截业务层的异常
				if r := recover(); r != nil {
					cerr = code.New(100000000, "Service internal error")
					serviceCtx.WithFields(logrus.Fields{
						"error": r,
						"stack": string(debug.Stack()),
					}).Errorln("Panic")
				}
				// 错误的返回
				if cerr != nil {
					httpCtx.JSON(200, map[string]interface{}{
						"result":  false,
						"mcode":   fmt.Sprintf("%s_%d", Prefix, cerr.Code()),
						"message": cerr.Error(),
					})
				} else {
					httpCtx.JSON(200, map[string]interface{}{
						"result": true,
						"data":   data,
					})
				}
				// 正确的返回

				l := serviceCtx.WithFields(logrus.Fields{
					"method": httpCtx.Request.Method,
					"path":   httpCtx.Request.URL.Path,
					"delay":  time.Since(since),
				})
				if cerr != nil {
					l.WithFields(logrus.Fields{
						"mcode":   fmt.Sprintf("%s_%d", Prefix, cerr.Code()),
						"message": cerr.Error(),
					}).Error("Http request failed")
				} else {
					l.Info("HTTP request done")
				}
			}()

			data, cerr = f(serviceCtx, httpCtx)
		}
	}

	go func() {
		router := gin.New()
		router.GET("/v1/any", wrapFunc(handler))
		fmt.Println(router.Run(":8080"))
	}()

	time.Sleep(time.Millisecond * 200)

	ret := map[string]interface{}{}
	invoke.Addr("127.0.0.1:8080").
		Get("/v1/any").
		Query("name", "xiaoming").
		Query("type", "1").
		Exec(&ret)

	invoke.Addr("127.0.0.1:8080").
		Get("/v1/any").
		Query("panic", "yes").Exec(nil)

	invoke.Addr("127.0.0.1:8080").
		Get("/v1/any").
		Query("error", "yes").Exec(nil)
	time.Sleep(time.Second * 5)
}
