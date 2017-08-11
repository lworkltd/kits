service.Context
-----------
集业务日志，Tracing于一体的上下文

功能
-----
1. 继承`Context`的功能
2. 丰富`logrus`日志的功能，并且可以附加服务器信息，日志的文件和行号和Tracing调用链信息
3. 从HTTP请求解析Tracing和将Tracing信息注入后继请求
4. 对于严重日志，会把日志内容推送一份给OpenTracing，异常日志中
用法
--------
`A服务`是一个HTTP API服务器，接受来自客户端的请求，当请求来临时，首先会先到`数据库`取出数据，然后再向`服务B`请求余下的请求数据:

```
func pullDataFromServiceB(ctx Context, dbData interface{}, parameters ...interface{}) (interface{}, code.Error) {
    url := "/any/path"
    service := "service-a"
    serviceId := "service-a-1"
    request, _ := http.NewRequest("GET", url, nil)

    subName := fmt.Sprintf("http://%s-%s:%s", service, serviceId, url)
    subContext := ctx.SubContext(subName)
    defer subContext.Finish()

    ctx.Inject(request.Header)
    client := http.Client{}
    client.Do(request)

    // parse data from response ...

    return "anydata", nil
}

func readDataFromDatabase(ctx Context, parameters ...interface{}) (interface{}, code.Error) {
    dbContext := ctx.SubContext("read-data-base")
    defer dbContext.Finish()

    // Read data from database ...

    return "anydata", nil
}

func readData(ctx Context, parameters ...interface{}) (interface{}, code.Error) {
    data, cerr := readDataFromDatabase(ctx, parameters...)
    if cerr != nil {
        logrus.WithFields(logrus.Fields{
            "parameters": parameters,
            "error":      cerr,
        }).Error("Read data base failed")
        return nil, cerr
    }

    return pullDataFromServiceB(ctx, data, parameters...)
}

func handler(serviceCtx Context, r *gin.Context) (interface{}, code.Error) {
    return readData(serviceCtx, r.Query("name"), r.Query("type"))
}

func wrapFunc(f func(Context, *gin.Context) (interface{}, code.Error)) func(*gin.Context) {
    return func(httpCtx *gin.Context) {
        Prefix := "SERVICE_A"
        logger := logrus.New()
        logger.SetLevel(logrus.WarnLevel)
        logger.Formatter = &logrus.JSONFormatter{}
        logger.Hooks.Add(logutils.NewServiceTagHook("service-a", "service-a-10", "dev"))
        logger.Hooks.Add(logutils.NewFileLineHook(true))
        serviceCtx, _ := FromHttpRequest(httpCtx.Request, logger)
        defer serviceCtx.Finish()

        var (
            data interface{}
            cerr code.Error
        )
        defer func() {
            if r := recover(); r != nil {
                cerr = code.New(100000000, "Service internal error")
                serviceCtx.WithFields(logrus.Fields{
                    "error": r,
                    "stack": string(debug.Stack()),
                }).Errorln("Panic")
            }

            httpCtx.JSON(200, map[string]interface{}{
                "result":  cerr == nil,
                "mcode":   fmt.Sprintf("%s_%d", Prefix, cerr.Code()),
                "message": cerr.Error(),
                "data":    data,
            })
        }()

        data, cerr = f(serviceCtx, httpCtx)
    }
}

func main(){
	router := gin.New()
	router.GET("/v1/any", wrapFunc(handler))
	router.Run(":8080")
}

invoke.Addr("127.0.0.1:8080").
    Get("/v1/any").
    Query("name", "xiaoming").
    Query("type", "1").
    Exec(&ret)

invoke.Addr("127.0.0.1:8080").
    Get("/v1/any").
    Query("panic", "yes").Exec(nil)
    
invoke.Addr("127.0.0.1:8080").
    Get("/v1/any").
    Query("error", "yes").Exec(nil)

```
