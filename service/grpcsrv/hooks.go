package grpcsrv

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
)

const (
	// DebugMode 调试模式
	DebugMode string = "debug"
	// ReleaseMode 发布模式
	ReleaseMode string = "release"
)

var (
	// MinWarningDelay 延迟告警时延
	MinWarningDelay = time.Second * 4
)

// HandlerFunc GRPC处理函数
type HandlerFunc func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse)

// HookFunc 钩子函数
type HookFunc func(f HandlerFunc) HandlerFunc

// HookRecover 当handler在收到panic时，能够恢复
func HookRecover(f HandlerFunc) HandlerFunc {
	return func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
		defer func() {
			if r := recover(); r != nil {
				cerr, is := r.(code.Error)
				if is {
					// 此错误一般由间接调用参数或一些通用错误残生
					if cerr.Mcode() != "" {
						commRsp = newErrorRsp(cerr.Mcode(), cerr.Error())
						return
					}
					// 此类错误一般由服务内部参数，返回了一个数字类型的错误码
					commRsp = newErrorRsp(
						fmt.Sprintf("%s%d", mcodePrefix, cerr.Code()),
						cerr.Error())

					return
				}

				// 服务内部错误，由服务内部出现Crash产生
				commRsp = newErrorRsp("GRPC_INTERNAL_ERROR", "grpc server internal error")

				// TODO:上报内容并通知开发人员处理异常
				// ReportPanic(exeFileName,stack())
				fmt.Println(r)
				fmt.Println(string(debug.Stack()))
				return
			}

		}()

		return f(ctx, commReq)
	}
}

// HookLogger 日志钩子
func HookLogger(f HandlerFunc) HandlerFunc {
	return func(ctx context.Context, commReq *grpccomm.CommRequest) *grpccomm.CommResponse {
		since := time.Now().UTC()
		r := f(ctx, commReq)
		costTime := time.Now().Sub(since)

		log := logrus.WithFields(logrus.Fields{
			"latency": costTime.String(),
			"reqName": commReq.ReqInterface,
		})

		if commReq.ReqSercice != "" {
			log = log.WithField("reqService", commReq.ReqSercice)
		}

		// 错误返回
		if r != nil && r.Result == false {
			log.WithFields(logrus.Fields{
				"mcode":   r.Mcode,
				"message": r.Message,
			}).Error("GRPC FAILED")
			return r
		}

		// 高延迟返回
		if costTime >= MinWarningDelay {
			log.Warn("GRPC LATENCY")
			return r
		}

		// 正常返回
		log.Info("GRPC DONE")

		return r
	}
}

// DefaultHooks 默认的钩子列表
var DefaultHooks = []HookFunc{
	HookLogger,
	HookRecover,
}

// DefenceSlowSideHook 防止雪崩
func DefenceSlowSideHook(n int32) HookFunc {
	return func(f HandlerFunc) HandlerFunc {
		return func(ctx context.Context, commReq *grpccomm.CommRequest) (commRsp *grpccomm.CommResponse) {
			err := checkSnowSlide(n)
			if err != nil {
				return newRspFromError(err)
			}
			return f(ctx, commReq)
		}
	}
}

var (
	usingHooks []HookFunc
)

// UseHook 使用钩子列表,靠前的钩子最先进入
func UseHook(hooks ...HookFunc) {
	usingHooks = append(usingHooks, hooks...)
}
