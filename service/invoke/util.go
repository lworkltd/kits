package invoke

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

func makeUrl(sche, host, path string, querys map[string][]string) (string, error) {
	b := make([]byte, 0, len(sche)+len(host)+len(path)+3+len(querys)*10)
	buffer := bytes.NewBuffer(b)

	buffer.WriteString(sche)
	buffer.WriteString("://")
	buffer.WriteString(host)
	buffer.WriteString(path)

	qv := url.Values{}
	for key, array := range querys {
		for _, value := range array {
			qv.Add(key, value)
		}
	}

	queryString := qv.Encode() // more performace? like appendEncode(&buffer,qv)

	if queryString != "" {
		buffer.WriteByte('?')
		buffer.WriteString(queryString)
	}

	return buffer.String(), nil
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
