package svc

import (
	"bytes"
	"net/url"
	"strings"
)

func parsePath(path string, r map[string]string) (string, error) {
	ret := path
	//FIXME:bad perpormance
	for key, value := range r {
		old := "{" + key + "}"
		ret = strings.Replace(ret, old, value, 1)
	}

	return ret, nil
}

func createURL(path string, querys map[string][]string) (string, error) {
	qv := url.Values{}
	for key, array := range querys {
		for _, value := range array {
			qv.Add(key, value)
		}
	}

	queryString := qv.Encode()
	if queryString != "" {
		path = path + "?" + queryString
	}

	return path, nil
}

func cutBytes(body []byte, max int) []byte {
	if max < 3 {
		panic("max parameter too small")
	}
	if len(body) < max {
		return body
	}
	copy(body[:max-3], bytes.Repeat([]byte{'.'}, 3))

	return body[:max]
}
