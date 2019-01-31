package httpsrv

import (
	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/httpsrv/httpstat"
)

// GroupWrapper 组封装
type GroupWrapper interface {
	BasePath() string
	Get(path string, f interface{})
	Patch(path string, f interface{})
	Post(path string, f interface{})
	Put(path string, f interface{})
	Options(path string, f interface{})
	Head(path string, f interface{})
	Delete(path string, f interface{})
	Any(path string, f interface{})
	Handle(method, path string, f interface{})

	HandlePprof()

	Group(path string) GroupWrapper
}

type groupWrapper struct {
	*gin.RouterGroup
	wrapper *Wrapper
}

func (gwrapper *groupWrapper) BasePath() string {
	return gwrapper.RouterGroup.BasePath()
}

func (gwrapper *groupWrapper) Group(path string) GroupWrapper {
	return &groupWrapper{
		RouterGroup: gwrapper.RouterGroup.Group(path),
		wrapper:     gwrapper.wrapper,
	}
}

func (gwrapper *groupWrapper) HandlePprof() {
	debugPrintRoute(gwrapper.wrapper.logger, "PPROF", gwrapper.RouterGroup.BasePath()+"/debug/pprof/.*", nil)
	ginpprof.WrapGroup(gwrapper.RouterGroup)
}

func (gwrapper *groupWrapper) HandleStat() {
	gwrapper.wrapper.enableStat = true
	gwrapper.Get("/debug/httpstat/delay", httpstat.StatDelay)
	gwrapper.Get("/debug/httpstat/result", httpstat.StatResult)
}

func (gwrapper *groupWrapper) Handle(method, path string, f interface{}) {
	debugPrintRoute(gwrapper.wrapper.logger, method, path, f)
	gwrapper.RouterGroup.Handle(method, path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Get(path string, f interface{}) {
	gwrapper.Handle("GET", path, f)
}

func (gwrapper *groupWrapper) Patch(path string, f interface{}) {
	gwrapper.Handle("PATCH", path, f)
}

func (gwrapper *groupWrapper) Post(path string, f interface{}) {
	gwrapper.Handle("POST", path, f)
}

func (gwrapper *groupWrapper) Put(path string, f interface{}) {
	gwrapper.Handle("PUT", path, f)
}

func (gwrapper *groupWrapper) Options(path string, f interface{}) {
	gwrapper.Handle("OPTIONS", path, f)
}

func (gwrapper *groupWrapper) Head(path string, f interface{}) {
	gwrapper.Handle("HEAD", path, f)
}

func (gwrapper *groupWrapper) Delete(path string, f interface{}) {
	gwrapper.Handle("DELETE", path, f)
}

func (gwrapper *groupWrapper) Any(path string, f interface{}) {
	debugPrintRoute(gwrapper.wrapper.logger, "ANY", path, f)
	gwrapper.Any(path, gwrapper.wrapper.wrapFunc(f))
}
