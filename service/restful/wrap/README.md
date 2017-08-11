WRAP
--------------------
本包提供了对进行LEANWORK的restful的协议封装  
具体要求参见：[FIXME](#)
Http框架采用的是gin

为什么不用Middleware
----
```
func(c *gin.Context) (interface{}, code.Error){
    // ...
}
```
1. Wrapper限定了输出的结果,让代码变得更加简洁,Handler不再关心数据流的写入，只需要返回数据和错误
2. 禁止HTTP设置为除200以外的状态，而错误都由Mcode来传递，让所有错误都具有可传递和可解释的特性，使用原生的，则需要重写Recover的中间件把500修改成为200，再返回一个代表内部出错的错误。
3. 限定了返回参数的Error，因此限定了如果程序发生错误，则必须返回错误码和错误原因；如果使用原生的，则每个错误的分支的都需要打包成为标准结构，累赘并且容易出错
4. 既然所有的Handler都返回了错误码，那么HTTP的AccessLog就应该增加错误码和错误原因来提高查找问题的效率，如果使用gin的中间件，则需要从数据流里面解析出错误码和错误原因再打印，这不怎么效率
5. Wrapper集成了Tracing，如果使用原生的，那么也需要重写一个中间件来实现这个功能

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

wrapper := New(&Option{
    Prefix: "MYSERVICE_EXCEPTION_",
})
foo := func(c *gin.Context) (interface{}, code.Error) {
    routeError := c.Params.ByName("error")
    if routeError == "yes" {
        return nil, code.New(FailedCode, "Foo failed!")
    }

    ret := &Data{
        Name: "Anna",
        Age:  15,
    }
    return ret, nil
}

bar := func(c *gin.Context) (interface{}, code.Error) {
    routeError := c.Params.ByName("error")
    if routeError == "yes" {
        return nil, code.New(FailedCode, "Bar failed!")
    }

    return &Data{
        Name: "Petter",
        Age:  32,
    }, nil
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
wrapper.Delete(v2, "/bar", bar)
```