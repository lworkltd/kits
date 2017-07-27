package invoke

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/golang/protobuf/proto"
	"github.com/opentracing/opentracing-go"
)

type client struct {
	service      Service
	path         string
	createTime   time.Time
	errInProcess error

	method   string
	host     string
	sche     string
	serverid string

	headers map[string]string
	querys  map[string][]string
	routes  map[string]string
	payload func() ([]byte, error)

	logFields  map[string]interface{}
	ctx        context.Context
	useTracing bool
	useCircuit bool
	fallback   func(error) error
}

func (client *client) circuitName() string {
	return client.serverid
}

func (client *client) tracingName() string {
	return client.serverid
}

func (client *client) clear() {
	client.service = nil
	client.path = ""
	client.createTime = time.Unix(0, 0)
	client.errInProcess = nil
	client.headers = nil
	client.routes = nil
	client.querys = nil
	client.method = "GET"
	client.host = ""
	client.sche = "http"
	client.payload = nil
	client.logFields = make(map[string]interface{}, 10)
	client.ctx = nil
}

func (client *client) Tls() Client {
	if client.errInProcess != nil {
		return client
	}

	client.sche = "https"

	return client
}

func (client *client) Fallback(func(error) error) Client {
	if client.errInProcess != nil {
		return client
	}
	return client
}

func (client *client) Header(headerName, headerValue string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.headers == nil {
		client.headers = map[string]string{headerName: headerValue}
		return client
	}

	client.headers[headerName] = headerValue

	return client
}

func (client *client) Headers(headers map[string]string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.headers == nil {
		client.headers = make(map[string]string, len(headers))
	}

	for key, value := range headers {
		client.headers[key] = value
	}

	return client
}

func (client *client) Query(queryName, queryValue string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.querys == nil {
		client.querys = map[string][]string{
			queryName: []string{queryValue},
		}
		return client
	}

	querys := client.querys[queryName]
	querys = append(querys, queryValue)
	client.querys[queryName] = querys

	return client
}

func (client *client) QueryArray(queryName string, queryValues ...string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.querys == nil {
		client.querys = map[string][]string{
			queryName: queryValues,
		}
		return client
	}

	querys := client.querys[queryName]
	querys = append(querys, queryValues...)
	client.querys[queryName] = querys

	return client
}

func (client *client) Querys(queryValues map[string][]string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.querys == nil {
		client.querys = make(map[string][]string, len(queryValues))
		return client
	}

	for key, qs := range queryValues {
		querys := client.querys[key]
		querys = append(querys, qs...)
		client.querys[key] = querys
	}

	return client
}

func (client *client) Route(routeName, routeTo string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.routes == nil {
		client.routes = map[string]string{routeName: routeTo}
		return client
	}

	client.routes[routeName] = routeTo

	return client
}

func (client *client) Routes(routes map[string]string) Client {
	if client.errInProcess != nil {
		return client
	}

	if client.routes == nil {
		client.routes = make(map[string]string, len(routes))
	}

	for routeName, route := range routes {
		client.routes[routeName] = route
	}

	return client
}

func (client *client) Json(payload interface{}) Client {
	if client.errInProcess != nil {
		return client
	}

	client.payload = func() ([]byte, error) {
		return json.Marshal(payload)
	}

	return client
}

func (client *client) Proto(payload proto.Message) Client {
	if client.errInProcess != nil {
		return client
	}

	client.payload = func() ([]byte, error) {
		return proto.Marshal(payload)
	}

	return client
}

func (client *client) Context(ctx context.Context) Client {
	if client.errInProcess != nil {
		return client
	}

	client.ctx = ctx

	return client
}

func (client *client) Exec(out interface{}) (int, error) {
	if client.useTracing {
		span, ctx := opentracing.StartSpanFromContext(client.ctx, client.tracingName())
		client.ctx = ctx
		defer span.Finish()
	}

	var err error
	var status int
	if !client.useCircuit {
		status, err = client.exec(out)
	} else {
		err = hystrix.Do(client.circuitName(), func() error {
			s, err := client.exec(out)
			status = s
			return err
		}, client.fallback)
	}

	if doLogger {
		fileds := logrus.Fields{
			"service":    client.service,
			"service_id": client.serverid,
			"method":     client.method,
			"path":       client.path,
			"querys":     client.querys,
			"headers":    client.headers,
			"routes":     client.routes,
			"endpoint":   client.host,
			"cost":       time.Since(client.createTime),
		}

		if err != nil {
			logrus.WithFields(fileds).WithError(err).Error("Invoke service failed")
		} else {
			logrus.WithFields(fileds).Error("Invoke service done")
		}
	}

	return status, err
}

func (client *client) build() (*http.Request, error) {
	path, err := parsePath(client.path, client.routes)
	if err != nil {
		client.logFields["error"] = "resources parameter invalid"
		client.logFields["routes"] = client.routes
		return nil, err
	}

	if client.host == "" {
		client.logFields["error"] = "no avaliable remote"
		return nil, fmt.Errorf("remote is emtpy")
	}

	url, err := makeUrl(client.sche, client.host, path, client.querys)

	if err != nil {
		client.logFields["scheme"] = client.sche
		client.logFields["remote"] = client.host
		client.logFields["path"] = url
		client.logFields["querys"] = client.querys
		return nil, err
	}

	reader := &bytes.Reader{}
	if client.payload != nil {
		b, err := client.payload()
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}

	request, err := http.NewRequest(client.method, url, reader)
	if err != nil {
		client.logFields["error"] = err
		return nil, fmt.Errorf("create http request failed,%v", err)
	}

	if client.ctx != nil {
		request.WithContext(client.ctx)
	}

	request.Header.Add("Content-Type", "application/json")

	for k, v := range client.headers {
		request.Header.Add(k, v)
	}

	return request, nil
}

func (client *client) exec(out interface{}) (int, error) {
	if client.errInProcess != nil {
		return 0, client.errInProcess
	}

	cli := &http.Client{}
	request, err := client.build()
	if err != nil {
		return 0, err
	}

	resp, err := cli.Do(request)
	if err != nil {
		client.logFields["error"] = err
		return 0, err
	}

	client.logFields["status"] = resp.StatusCode
	client.logFields["status_code"] = resp.Status

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		client.logFields["error"] = "response status error"
		return resp.StatusCode, fmt.Errorf("reponse with bad status,%d", resp.StatusCode)
	}

	rsp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		client.logFields["error"] = err
		return resp.StatusCode, fmt.Errorf("Read response body failed")
	}
	defer resp.Body.Close()

	client.logFields["response_payload_len"] = len(rsp)

	err = json.Unmarshal(rsp, out)
	if err != nil {
		client.logFields["error"] = err
		client.logFields["content"] = string(cutBytes(rsp, 4096))
		return 0, fmt.Errorf("marshal result body failed")
	}

	return resp.StatusCode, nil
}
