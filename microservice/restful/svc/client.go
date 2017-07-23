package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/golang/protobuf/proto"
	"github.com/opentracing/opentracing-go"
)

type client struct {
	service      IService
	path         string
	createTime   time.Time
	errInProcess error

	method string
	host   string
	sche   string

	headers map[string]string
	routes  map[string]string
	querys  map[string][]string
	payload func() ([]byte, error)

	logFields  map[string]interface{}
	ctx        context.Context
	useTracing bool
	useCircuit bool
}

func (client *client) circuitName() string {
	return client.service.Name() + "/" + client.path
}

func (client *client) tracingName() string {
	return client.service.Name() + "/" + client.path
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

func (client *client) Exec(out interface{}) error {
	if client.useTracing {
		span, ctx := opentracing.StartSpanFromContext(client.ctx, client.tracingName())
		client.ctx = ctx
		defer span.Finish()
	}

	if !client.useCircuit {
		return client.exec(out)
	}

	return hystrix.Do(client.circuitName(), func() error {
		return client.exec(out)
	}, nil)
}

func (client *client) exec(out interface{}) error {
	if client.errInProcess != nil {
		return client.errInProcess
	}

	path, err := parsePath(client.path, client.routes)
	if err != nil {
		client.logFields["error"] = "resources parameter invalid"
		client.logFields["routes"] = client.routes
		return err
	}

	if client.host == "" {
		client.logFields["error"] = "no avaliable remote"
		return fmt.Errorf("Remote is emtpy")
	}

	url, err := makeUrl(client.sche, client.host, path, client.querys)

	if err != nil {
		client.logFields["scheme"] = client.sche
		client.logFields["remote"] = client.host
		client.logFields["path"] = url
		client.logFields["querys"] = client.querys
		return err
	}

	var reader *bytes.Reader
	if client.payload != nil {
		b, err := client.payload()
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	request, err := http.NewRequest(client.method, url, reader)
	if err != nil {
		client.logFields["error"] = err
		return fmt.Errorf("Create HTTP request failed")
	}

	if client.ctx != nil {
		request.WithContext(client.ctx)
	}

	request.Header.Add("Content-Type", "application/json")

	for k, v := range client.headers {
		request.Header.Add(k, v)
	}

	cli := &http.Client{}

	resp, err := cli.Do(request)
	if err != nil {
		client.logFields["error"] = err
		return err
	}

	client.logFields["status"] = resp.StatusCode
	client.logFields["status"] = resp.Status

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		client.logFields["error"] = "response status error"
		return fmt.Errorf("Reponse with bad status")
	}

	rsp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		client.logFields["error"] = err
		return fmt.Errorf("Read response body failed")
	}
	defer resp.Body.Close()

	client.logFields["response_payload_len"] = len(rsp)

	err = json.Unmarshal(rsp, out)
	if err != nil {
		client.logFields["error"] = err
		client.logFields["content"] = string(cutBytes(rsp, 4096))
		return fmt.Errorf("Marshal result body failed")
	}

	return nil
}
