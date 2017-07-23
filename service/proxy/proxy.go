package proxy

import (
	"net/http"
)

// Proxy 提供代理服务
type Proxy interface {
	Proxy(string, string, func(*http.Request, *http.Response))
}
