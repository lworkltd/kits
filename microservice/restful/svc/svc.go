package svc

import (
	"time"

	"github.com/golang/protobuf/proto"
)

// DiscoveryFunc ...
type DiscoveryFunc func(name string) ([]string, error)

// IEngine 引擎
type IEngine interface {
	Service(string) IService
	SetDiscovery(DiscoveryFunc)
}

// IService 服务
type IService interface {
	Get(string) IClient
	Post(string) IClient
	Put(string) IClient
	Delete(string) IClient
	Method(string, string) IClient
	Name() string
}

// IClient 客户端
type IClient interface {
	Whole(interface{}) IClient
	Header(map[string]string) IClient
	Query(map[string][]string) IClient
	Route(map[string]string) IClient
	Json(interface{}) IClient
	Timeout(time.Duration) IClient
	Request(interface{}) error
	Proto(proto.Message) IClient
}

// Whole
// 声明：其实不建议使用这种方法，因为始终无法避免计算reflect.Value(v)
type Whole struct {
	HeaderParams interface{}
	PathParams   interface{}
	QueryParams  interface{}
	Payload      interface{}
}

var engine IEngine = newEngine()

// SetDiscovery 设置服务发现
// 如果没有设置则会返回 ErrDiscoveryNotConfig 错误
func SetDiscovery(f DiscoveryFunc) {
	engine.SetDiscovery(f)
}

// Service 返回服务器
// 也是服务调用的入口
// eg:
// svc.Service("my-service").Get("/v1/fruits/{fruit}/weight").
// 	Header(map[string]string{"Any-Header": "AnyHeaderValue"}).
// 	Query(map[string][]string{
// 		"season": {"summer", "spring"},
// 		"page":   {"1"},
// 	}).
// 	Object(map[string]string{
// 		"fruit": "apple",
// 	}).
// 	Request(&a)
// More examples please see <<README.md>>
func Service(name string) IService {
	return engine.Service(name)
}
