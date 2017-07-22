package svc

import (
	"reflect"
	"testing"
)

func TestClient(t *testing.T) {
	var a map[string]interface{}
	Service("my-service").Get("/v1/fruits/{fruit}/weight").
		Header(map[string]string{"Any-Header": "AnyHeaderValue"}).
		Query(map[string][]string{
			"season": {"summer", "spring"},
			"page":   {"1"},
		}).
		Object(map[string]string{
			"fruit": "apple",
		}).
		Request(&a)

}

type FeignKey struct {
	Product string `json:"productId"`
	Tenant  string `json:"tenantId"`
}

type HeaderItem struct {
	F1          float64   `header:"X-COUNT"`
	FeignKey1   FeignKey  `header:"Feign-Key1,json"`
	FeignKey2   *FeignKey `header:"Feign-Key2,json"`
	ContentType string    `header:"Content-Type"`
	Gzip        []string  `header:"gzip,comma"`
	Fruits      []string  `header:"Fruits,json"`
}

type FormItem struct {
	Planets   []string `query:"planet,comma"`
	Places    []string `query:"planet"`
	Amount    float64  `query:"amount"`
	Ignore    float64  `query:"-"`
	Timestamp int64    `query:"timestamp"`
	Count     int32    `query:"count"`
	Days      int16    `query:"day"`
	Words     int8     `query:"words"`
}

type Route struct {
	Country string `route:"country"`
	City    string `route:"city"`
	PostNo  int    `route:"positno"`
}

func TestHeader(t *testing.T) {
	header := &HeaderItem{
		F1:          0.000001,
		FeignKey1:   FeignKey{"TW", "T0001111"},
		FeignKey2:   &FeignKey{"FW", "T0002222"},
		ContentType: "我是电费一⒈ ds",
		Gzip:        []string{"gzip", "123", ".1234"},
		Fruits:      []string{"apple", "orange"},
	}

	expect := map[string][][]byte{
		"X-COUNT":      {[]byte("0.000001")},
		"Feign-Key1":   {[]byte(`{"productId":"TW","tenantId":"T0001111"}`)},
		"Feign-Key2":   {[]byte(`{"productId":"TW","tenantId":"T0002221"}`)},
		"Content-Type": {[]byte("我是电费一⒈ ds")},
		"Gzip":         {[]byte("gizp"), []byte("123"), []byte(".1234")},
		"Fruits":       {[]byte(`["apple","orange"]`)},
	}

	bs, err := Marshal(header, "header")
	if err != nil {
		panic(err)
	}
	for k, e := range expect {
		v, exist := bs[k]
		if !exist {
			panic("key " + k + " not exist")
		}
		if !reflect.DeepEqual(e, v) {
			t.Errorf("key %s not equal expect %s got %s", k, e, v)
		}
	}
}

func BenchmarkReflectValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reflect.ValueOf(HeaderItem{})
	}
}
