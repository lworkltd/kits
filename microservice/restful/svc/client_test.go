package svc

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func Test_client_exec(t *testing.T) {
	type Response struct {
		ResultCode int `json:"result_code"`
	}
	type Reqeust struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	go func() {
		http.HandleFunc("/v1/country/china/city/chengdu/street/longjiangroad", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				w.WriteHeader(500)
				return
			}
			if r.FormValue("hourse_number") != "T-12" {
				w.WriteHeader(500)
				return
			}
			if r.FormValue("building") != "12" {
				w.WriteHeader(500)
				return
			}
			if r.FormValue("floor") != "1" {
				w.WriteHeader(500)
				return
			}
			if r.FormValue("room") != "1" {
				w.WriteHeader(500)
				return
			}
			if r.Header.Get("Registration-Personnel") != "anna.liu" {
				w.WriteHeader(500)
				return
			}
			if r.Header.Get("statistical-auth-code") != "023432" {
				w.WriteHeader(500)
				return
			}
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			defer r.Body.Close()

			person := &Reqeust{}
			err = json.Unmarshal(b, person)
			if person.Name != "小华" {
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
			rb, _ := json.Marshal(&Response{
				ResultCode: 200,
			})
			w.Write(rb)
		})
		http.ListenAndServe(":26403", nil)
	}()

	time.Sleep(time.Millisecond * 100)

	service := &ServiceImpl{
		discover: func(string) ([]string, error) {
			return []string{"127.0.0.1:26403"}, nil
		},
		name: "test-service",
	}
	var response Response
	_, err := service.Post("/v1/country/{country}/city/{city}/street/{street}").
		Route("street", "longjiangroad").
		Routes(map[string]string{
			"country": "china",
			"city":    "chengdu",
		}).
		Query("hourse_number", "T-12").
		QueryArray("building", "12", "15").
		Querys(map[string][]string{
			"floor": []string{"1", "2", "3"},
			"room":  []string{"1"},
		}).
		Header("job", "student").
		Headers(map[string]string{
			"Registration-Personnel": "anna.liu",
			"statistical-auth-code":  "023432",
		}).
		Json(&Reqeust{Name: "小华", Age: 123}).
		Context(context.Background()).Exec(&response)
	if err != nil {
		t.Errorf("client.exec() error = %v", err)
		return
	}

	if response.ResultCode != 200 {
		t.Errorf("client.exec() ResultCode = %v expect %d", response.ResultCode, 200)
	}
}
