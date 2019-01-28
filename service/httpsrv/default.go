package httpsrv

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/sirupsen/logrus"
)

// defaultWrap 默认的函数转换
func (wrapper *Wrapper) defaultWrap(fx interface{}) gin.HandlerFunc {
	f := fx.(func(ginCtx *gin.Context) (interface{}, code.Error))
	return func(httpCtx *gin.Context) {
		Prefix := wrapper.mcodePrefix // 错误码前缀
		since := time.Now()
		var (
			data interface{}
			cerr code.Error
		)

		defer func() {
			// 拦截业务层的异常
			if r := recover(); r != nil {
				fmt.Println(r)
				if codeErr, ok := r.(code.Error); ok {
					cerr = codeErr
				} else {
					fmt.Println("Panic", r)
					fmt.Println(string(debug.Stack()))
					cerr = code.NewMcode("SERVICE_INTERVAL_ERROR", "Service internal error")
				}
			}

			l := wrapper.logger.WithFields(logrus.Fields{
				"method": httpCtx.Request.Method,
				"path":   httpCtx.Request.URL.Path,
				"delay":  time.Since(since),
			})

			// 错误的返回
			if cerr != nil {
				if cerr.Mcode() != "" {
					httpCtx.JSON(http.StatusOK, map[string]interface{}{
						"result":    false,
						"mcode":     cerr.Mcode(),
						"message":   cerr.Message(),
						"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
					})

					l = l.WithFields(logrus.Fields{
						"mcode": cerr.Mcode(),
					})
				} else {
					mcode := fmt.Sprintf("%s_%d", Prefix, cerr.Code())
					httpCtx.JSON(http.StatusOK, map[string]interface{}{
						"result":    false,
						"mcode":     mcode,
						"message":   cerr.Message(),
						"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
					})

					l = l.WithFields(logrus.Fields{
						"mcode": mcode,
					})
				}
			} else {
				resp := map[string]interface{}{
					"result":    true,
					"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
				}
				if data != nil {
					resp["data"] = data
				}
				httpCtx.JSON(http.StatusOK, resp)
			}

			if cerr != nil {
				l.WithFields(logrus.Fields{
					"message": cerr.Message(),
				}).Error("HTTP request failed")
			} else {
				l.Info("HTTP request done")
			}
		}()

		// 过载保护
		if wrapper.snowSlide != nil {
			cerr = wrapper.snowSlide.Check(httpCtx)
			if cerr == nil {
				data, cerr = f(httpCtx)
			}
		} else {
			data, cerr = f(httpCtx)
		}

		wrapper.report(cerr, httpCtx, since)
	}
}
