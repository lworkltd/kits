WRAP
--------------------
本包提供了对进行LEANWORK的restful的协议封装  
具体要求参见：[FIXME](#)
Http框架采用的是gin

结构
-----
```
{
    "result":{true|false}
    "mcode":"ANYPREFIX_{CODE}",
    "message":"{any description}",
    "data":{response object},
}
```

用法
---------

```
const{
    FailedCode = 10010    
}

type Data struct{
    Name string
    Age int
}

wrapper := wrap.NewWraper("MYSERVICE_EXCEPTION_")
func Foo(c *gin.Context) wrap.Response {
    routeError := c.Params.ByName("error")
    if routeError = "yes"{
        return wrapper.Errorf(
            FailedCode,
            "failed!routeError=%s",routeError,
        )
    }

    return &wrapper.Done(&Data{
        "Name":"Anna",
        "Age":123,
    })
}

func main() {
	r := gin.Default()
	wrapper.Get(r,"/request", wrapper.Wrap(Foo))
    v2 := router.Group("/v2")
	r.Run() // listen and serve on 0.0.0.0:8080
}
```