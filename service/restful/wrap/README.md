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
FailedCode := 10010

type Data struct {
    Name string
    Age  int
}

wrapper := New("MYSERVICE_EXCEPTION_")
foo := func(c *gin.Context) Response {
    routeError := c.Params.ByName("error")
    if routeError == "yes" {
        return wrapper.Errorf(FailedCode, "Foo failed! %v", routeError)
    }

    ret := &Data{
        Name: "Anna",
        Age:  15,
    }
    return wrapper.Done(ret)
}

bar := func(c *gin.Context) Response {
    routeError := c.Params.ByName("error")
    if routeError == "yes" {
        return wrapper.Error(FailedCode, "Bar Failed")
    }

    return wrapper.Done(&Data{
        Name: "Petter",
        Age:  32,
    })
}

r := gin.Default()
wrapper.Get(r, "/foo", foo)

v2 := r.Group("/v2")
wrapper.Post(v2, "/bar", bar)
wrapper.Get(v2, "/bar", bar)
wrapper.Put(v2, "/bar", bar)
wrapper.Options(v2, "/bar", bar)
wrapper.Patch(v2, "/bar", bar)
wrapper.Head(v2, "/bar", bar)
```