invoke
--------
服务调用


使用方法
------

```
import (
	"context"
	"fmt"
	"github.com/lworkltd/kits/service/invoke"
)

func main() {
	type Response struct {
		ResultCode int `json:"result_code"`
	}
	type Request struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	invoke.Init(&invoke.Option{
		Discover: func(name string) ([]string, error) {
			if name == "service-name" {
				return []string{"your-service-ip"}, nil
			}
			return nil, fmt.Errorf("service not found")
		},
		UseTracing: true,
		UseCircuit: true,
	})

	var response Response
	_, err := invoke.Name("service-name").Post("/v1/country/{country}/city/{city}/street/{street}").
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
		Json(&Request{Name: "小华", Age: 123}).
		Context(context.Background()).
		Exec(&response)
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
```