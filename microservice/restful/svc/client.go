package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
)

type Client struct {
	service      IService
	pathFormat   string
	createTime   time.Time
	errInProcess error
	headerParams map[string]string
	pathParams   map[string]string
	queryParams  map[string][]string
	method       string
	requestHost  string
	sche         string
	payload      interface{}
	protoPayload proto.Message
	logFields    map[string]interface{}
	timeout      time.Duration
}

func (client *Client) Header(header map[string]string) IClient {
	client.headerParams = header
	return client
}

func (client *Client) Query(query map[string][]string) IClient {
	client.queryParams = query
	return client
}

func (client *Client) Route(pathParams map[string]string) IClient {
	client.pathParams = pathParams
	return client
}

func (client *Client) Json(payload interface{}) IClient {
	client.payload = payload
	return client
}
func (client *Client) Proto(payload proto.Message) IClient {
	client.payload = payload
	return client
}

func (client *Client) Timeout(dur time.Duration) IClient {
	client.timeout = dur
	return client
}

func (client *Client) Context(ctx *context.Context) IClient {
	if client.errInProcess != nil {
		return client
	}

	return client
}

func (client *Client) Request(out interface{}) error {
	if client.errInProcess != nil {
		return client.errInProcess
	}

	path, err := parsePath(client.pathFormat, client.pathParams)
	if err != nil {
		client.logFields["error"] = "resources parameter invalid"
		client.logFields["resources"] = client.Object
		return err
	}

	if client.requestHost == "" {
		client.logFields["error"] = "no avaliable remote"
		return fmt.Errorf("Remote is emtpy")
	}

	url, err := createURL(
		fmt.Sprintf("%s://%s%s", client.sche, client.requestHost, path),
		client.queryParams,
	)

	if err != nil {
		client.logFields["scheme"] = client.sche
		client.logFields["remote"] = client.requestHost
		client.logFields["path"] = url
		client.logFields["query_args"] = client.queryParams
		return err
	}

	var reader *bytes.Reader
	if client.payload != nil {
		b, err := json.Marshal(client.payload)
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

	request.Header.Add("Content-Type", "application/json")
	for k, v := range client.headerParams {
		request.Header.Add(k, v)
	}

	if client.errInProcess != nil {
		return err
	}

	cli := &http.Client{
		Timeout: client.timeout,
	}
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

	err = json.Unmarshal(rsp, &out)
	if err != nil {
		client.logFields["error"] = err
		client.logFields["rsp_content"] = string(cutBytes(rsp, 1024))
		return fmt.Errorf("Marshal result body failed")
	}

	return nil
}

func (client *Client) Whole(a interface{}) IClient {
	return client
}
