package httpsrv

import (
	"fmt"

	"github.com/gin-gonic/gin"
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

func (gwrapper *groupWrapper) Handle(method, path string, f interface{}) {
	debugPrintRoute(method, path, f)
	gwrapper.RouterGroup.Handle(method, path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Get(path string, f interface{}) {
	gwrapper.RouterGroup.Handle("GET", path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Patch(path string, f interface{}) {
	gwrapper.RouterGroup.Handle("PATCH", path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Post(path string, f interface{}) {
	gwrapper.RouterGroup.Handle("POST", path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Put(path string, f interface{}) {
	gwrapper.RouterGroup.Handle("PUT", path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Options(path string, f interface{}) {
	gwrapper.RouterGroup.Handle("OPTIONS", path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Head(path string, f interface{}) {
	gwrapper.RouterGroup.Handle("HEAD", path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Delete(path string, f interface{}) {
	gwrapper.RouterGroup.Handle("DELETE", path, gwrapper.wrapper.wrapFunc(f))
}

func (gwrapper *groupWrapper) Any(path string, f interface{}) {
	gwrapper.RouterGroup.Any(path, gwrapper.wrapper.wrapFunc(f))
}

func debugPrintRoute(method, path string, f interface{}) {
	fmt.Printf("%12s%s%s", method, path, f)
}
