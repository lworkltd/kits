package monitor

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"
    "time"

    "github.com/Sirupsen/logrus"
    "net"
    "runtime"
)

const (
    reportMetricsDataInterval     = 3    // 上报监控数据的时间间隔，3秒
    reportRuntimeDataInterval     = 60   // 上报程序运行数据的时间间隔，60秒
    checkReportDataCountLimit     = 1000 // 当上报checkReportDataCountLimit记录后，检查一次是否该发送数据
    notReportDataSleepMillisecond = 10   // 没有上报数据时的休眠时间，单位毫秒，避免循环消耗太多CPU
    delimit                       = "#@#"
    defaultAliReportAddr          = "open.cms.aliyun.com"
    defaultStatsdReportAddr       = "metrics.lwork.com:9125"
    reportQueueLength             = 300 // 上报数据队列的长度
    sendReportDataTimeoutSecond   = 3   // 发送上报数据到阿里云的超时时间，单位为秒
)

// generatekey 简易序列化
func (me *ReqSuccessCountDimension) generatekey() string {
    return me.SName + delimit + me.SIP + delimit + me.TName + delimit + me.TIP + delimit + me.Infc
}

// parseSuccessCountDimension 简易反序列化
func parseSuccessCountDimension(successCountKey string) *ReqSuccessCountDimension {
    strArray := strings.Split(successCountKey, delimit)
    if 5 != len(strArray) {
        return nil
    }
    var obj ReqSuccessCountDimension
    obj.SName = strArray[0]
    obj.SIP = strArray[1]
    obj.TName = strArray[2]
    obj.TIP = strArray[3]
    obj.Infc = strArray[4]
    return &obj
}

func (me *ReqSuccessCountDimension) getMetricName() string {
    return "req_success_count"
}

// generatekey 简易序列化
func (me *ReqFailedCountDimension) generatekey() string {
    return me.SName + delimit + me.TName + delimit + me.TIP + delimit + me.Code + delimit + me.Infc
}

// parseFailedCountDimension 简易反序列化
func parseFailedCountDimension(failedCountKey string) *ReqFailedCountDimension {
    strArray := strings.Split(failedCountKey, delimit)
    if 5 != len(strArray) {
        return nil
    }
    var obj ReqFailedCountDimension
    obj.SName = strArray[0]
    obj.TName = strArray[1]
    obj.TIP = strArray[2]
    obj.Code = strArray[3]
    obj.Infc = strArray[4]
    return &obj
}

func (me *ReqFailedCountDimension) getMetricName() string {
    return "req_failed_count"
}

// generatekey 简易序列化
func (me *ReqSuccessAvgTimeDimension) generatekey() string {
    return me.SName + delimit + me.SIP + delimit + me.TName + delimit + me.TIP + delimit + me.Infc
}

// parseSuccessAvgTimeDimension 简易反序列化
func parseSuccessAvgTimeDimension(successAvgTimeKey string) *ReqSuccessAvgTimeDimension {
    strArray := strings.Split(successAvgTimeKey, delimit)
    if 5 != len(strArray) {
        return nil
    }
    var obj ReqSuccessAvgTimeDimension
    obj.SName = strArray[0]
    obj.SIP = strArray[1]
    obj.TName = strArray[2]
    obj.TIP = strArray[3]
    obj.Infc = strArray[4]
    return &obj
}
func (me *ReqSuccessAvgTimeDimension) getMetricName() string {
    return "req_success_avg_time"
}

// parseFailedAvgTimeDimension 简易序列化
func (me *ReqFailedAvgTimeDimension) generatekey() string {
    return me.SName + delimit + me.SIP + delimit + me.TName + delimit + me.TIP + delimit + me.Infc
}

// parseFailedAvgTimeDimension 简易反序列化
func parseFailedAvgTimeDimension(failedAvgTimeKey string) *ReqFailedAvgTimeDimension {
    strArray := strings.Split(failedAvgTimeKey, delimit)
    if 5 != len(strArray) {
        return nil
    }
    var obj ReqFailedAvgTimeDimension
    obj.SName = strArray[0]
    obj.SIP = strArray[1]
    obj.TName = strArray[2]
    obj.TIP = strArray[3]
    obj.Infc = strArray[4]
    return &obj
}

func (me *ReqFailedAvgTimeDimension) getMetricName() string {
    return "req_failed_avg_time"
}

func (me *RuntimeDataDimension) getMetricName() string {
    return "runtime_data"
}

// countInfo 用于计数
type countInfo struct {
    counter int64 // 次数
    sum     int64 // 总和，例如耗时总和（微秒）
}

type reqSuccessTimeConsumeInfo struct {
    succAvgTimeDimension *ReqSuccessAvgTimeDimension
    timeConsume          int64 // 耗时，单位微秒(1/1000000 秒）
}
type reqFailedTimeConsumeInfo struct {
    failedAvgTimeDimension *ReqFailedAvgTimeDimension
    timeConsume            int64 // 耗时，单位微秒(1/1000000 秒）
}

type monitorInfo struct {
    conf                     MonitorConf
    reqSuccCountChan         chan *ReqSuccessCountDimension  // 请求成功计数上报队列
    reqFailedCountChan       chan *ReqFailedCountDimension   // 请求失败计数上报队列
    reqSuccTimeConsumeChan   chan *reqSuccessTimeConsumeInfo // 请求成功耗时上报队列
    reqFailedTimeConsumeChan chan *reqFailedTimeConsumeInfo  // 请求失败耗时上报队列
    succCountMap             map[string]countInfo            // 请求成功次数计数，key为ReqSuccessCountDimension序列化字符串
    failedCountMap           map[string]countInfo            // 请求失败次数计数，key为ReqFailedCountDimension序列化后字符串
    succAvgTimeMap           map[string]countInfo            // 请求成功平均耗时计数，key为ReqSuccessAvgTimeDimension序列化字符串
    failedAvgTimeMap         map[string]countInfo            // 请求失败平均耗时计数，key为ReqFailedAvgTimeDimension序列化字符串
}

var (
    monitorObj          monitorInfo
    lastSendReport      time.Time
    lastSendRuntimeData time.Time
)

// checkAndReportData 检查是否需要发送上报数据，若需要则发送后返回并且修改lastSendToAliyunTime为当前时间
func (me *monitorInfo) checkAndReportData() bool {
    timeNow := time.Now()
    if timeNow.Unix()-lastSendReport.Unix() < reportMetricsDataInterval {
        return false
    }

    succCountMap := me.succCountMap
    failedCountMap := me.failedCountMap
    succAvgTimeMap := me.succAvgTimeMap
    failedAvgTimeMap := me.failedAvgTimeMap
    me.succCountMap = make(map[string]countInfo)
    me.failedCountMap = make(map[string]countInfo)
    me.succAvgTimeMap = make(map[string]countInfo)
    me.failedAvgTimeMap = make(map[string]countInfo)

    _, statsdRuntimeMetrics := getRuntimeMetrics(timeNow)  //获取程序运行状态信息
    if monitorObj.conf.EnableReport { //上报到阿里云
        succCountMetrics := getSuccCountAliyunMetrics(succCountMap, timeNow)
        failedCountMetrics := getFailedCountAliyunMetrics(failedCountMap, timeNow)
        succAvgMetrics := getSuccAvgTimeAliyunMetrics(succAvgTimeMap, timeNow)
        failedAvgMetrics := getFailedAvgTimeAliyunMetrics(failedAvgTimeMap, timeNow)
        aliMetrics := append(succCountMetrics, failedCountMetrics...)
        aliMetrics = append(aliMetrics, succAvgMetrics...)
        aliMetrics = append(aliMetrics, failedAvgMetrics...)
        //aliMetrics = append(aliMetrics, aliRuntimeMetrics...)  //去除metrics上报到阿里云，以减少时间序列数
        go sendRequestToAliMonitor(aliMetrics)
    }

    if monitorObj.conf.EnableStatsd { //上报到Statsd
        succCountMetrics := getSuccCountStatsdMetrics(succCountMap, timeNow)
        failedCountMetrics := getFailedCountStatsdMetrics(failedCountMap, timeNow)
        succAvgMetrics := getSuccAvgTimeStatsdMetrics(succAvgTimeMap, timeNow)
        failedAvgMetrics := getFailedAvgTimeStatsdMetrics(failedAvgTimeMap, timeNow)
        statsdMetrics := append(succCountMetrics, failedCountMetrics...)
        statsdMetrics = append(statsdMetrics, succAvgMetrics...)
        statsdMetrics = append(statsdMetrics, failedAvgMetrics...)
        statsdMetrics = append(statsdMetrics, statsdRuntimeMetrics...)
        go sendMetricsToStatsd(statsdMetrics)
    }

    lastSendReport = time.Now()
    return true
}

// processReportData 处理上报数据的函数
func (me *monitorInfo) processReportData() {
    reportCountTmp := 0
    lastSendReport = time.Now()

    for {
        select {
        case item := <-me.reqSuccCountChan:
            key := item.generatekey()
            countObj, exist := monitorObj.succCountMap[key]
            if false == exist {
                countObj = countInfo{counter: 0, sum: 0}
            }
            countObj.counter++
            monitorObj.succCountMap[key] = countObj
            reportCountTmp++
            if reportCountTmp > checkReportDataCountLimit && me.checkAndReportData() { // 避免上报数据太多，长时间没机会执行reportData
                reportCountTmp = 0
            }
        case item := <-me.reqFailedCountChan:
            key := item.generatekey()
            countObj, exist := monitorObj.failedCountMap[key]
            if false == exist {
                countObj = countInfo{counter: 0, sum: 0}
            }
            countObj.counter++
            monitorObj.failedCountMap[key] = countObj
            reportCountTmp++
            if reportCountTmp > checkReportDataCountLimit && me.checkAndReportData() { // 避免上报数据太多，长时间没机会执行reportData
                reportCountTmp = 0
            }
        case item := <-me.reqSuccTimeConsumeChan:
            key := item.succAvgTimeDimension.generatekey()
            countObj, exist := monitorObj.succAvgTimeMap[key]
            if false == exist {
                countObj = countInfo{counter: 0, sum: 0}
            }
            countObj.counter++
            countObj.sum += item.timeConsume
            monitorObj.succAvgTimeMap[key] = countObj
            reportCountTmp++
            if reportCountTmp > checkReportDataCountLimit && me.checkAndReportData() { // 避免上报数据太多，长时间没机会执行reportData
                reportCountTmp = 0
            }
        case item := <-me.reqFailedTimeConsumeChan:
            key := item.failedAvgTimeDimension.generatekey()
            countObj, exist := monitorObj.failedAvgTimeMap[key]
            if false == exist {
                countObj = countInfo{counter: 0, sum: 0}
            }
            countObj.counter++
            countObj.sum += item.timeConsume
            monitorObj.failedAvgTimeMap[key] = countObj
            reportCountTmp++
            if reportCountTmp > checkReportDataCountLimit && me.checkAndReportData() { // 避免上报数据太多，长时间没机会执行reportData
                reportCountTmp = 0
            }
        default:
            if me.checkAndReportData() { // 避免上报数据太多，长时间没机会执行reportData
                reportCountTmp = 0
            }
            time.Sleep(time.Millisecond * notReportDataSleepMillisecond) // 无上报数据时，休眠notReportDataSleepMillisecond毫秒，避免不断消耗CPU
        }
    }
}

// AliyunMetric 阿里云统计模型
type AliyunMetric struct {
    MetricName string      `json:"metricName"`
    Value      int64       `json:"value"`
    Timestamp  int64       `json:"timestamp"`
    Unit       string      `json:"unit"`
    Dimensions interface{} `json:"dimensions"`
}

//获取程序运行状态信息：协程数、堆数量、线程数、阻塞数、加锁数
func getRuntimeMetrics(timeNow time.Time) ([]AliyunMetric, []string) {
    aliMetrics := make([]AliyunMetric, 0)
    statsdMetrics := make([]string, 0)
    if timeNow.Unix()-lastSendRuntimeData.Unix() < reportRuntimeDataInterval {
        return aliMetrics, statsdMetrics
    }
    routineCount := runtime.NumGoroutine()
    threadCreateCount, _ := runtime.ThreadCreateProfile(nil)
    heapCount, _ := runtime.MemProfile(nil, true)
    blockCount, _ := runtime.BlockProfile(nil)
    mutexCount, _ := runtime.MutexProfile(nil)

    optionMap := make(map[string]int, 0)
    optionMap["routineCount"] = routineCount
    optionMap["threadCreateCount"] = threadCreateCount
    optionMap["heapCount"] = heapCount
    optionMap["blockCount"] = blockCount
    optionMap["mutexCount"] = mutexCount
    for option, value := range optionMap {
        var dimessionObj RuntimeDataDimension
        dimessionObj.SerName = monitorObj.conf.CurServiceName + "_" + monitorObj.conf.EnvironmentType
        dimessionObj.SerIP = monitorObj.conf.CurServerIP
        dimessionObj.Option = option

        var metric AliyunMetric
        metric.MetricName = dimessionObj.getMetricName()
        metric.Value = int64(value)
        metric.Timestamp = timeNow.UnixNano() / 1e6
        metric.Dimensions = dimessionObj
        aliMetrics = append(aliMetrics, metric)

        statsdMetric := fmt.Sprintf("runtime.data.%v.%v.%v:%v|g", strings.Replace(dimessionObj.SerName, ".", "-", -1), strings.Replace(dimessionObj.SerIP, ".", "-", -1), dimessionObj.Option, value)
        statsdMetrics = append(statsdMetrics, statsdMetric)
    }

    lastSendRuntimeData = timeNow
    return aliMetrics, statsdMetrics
}

func sendRequestToAliMonitor(metrics []AliyunMetric) error {
    if len(metrics) <= 0 {
        return nil
    }
    metricsBytes, _ := json.Marshal(metrics)
    body := fmt.Sprintf("userId=%v&namespace=%v&metrics=%v", monitorObj.conf.AliUid, monitorObj.conf.AliNamespace, string(metricsBytes))
    url := "http://" + monitorObj.conf.ReportAddr + "/metrics/put"
    request, err := http.NewRequest("POST", url, strings.NewReader(body))
    if err != nil {
        logrus.WithFields(logrus.Fields{"err": err, "url": url, "body": body}).Error("http.NewRequest failed")
        return err
    }
    request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    request.Header.Add("Connection", "close")

    cli := &http.Client{Timeout: time.Second * sendReportDataTimeoutSecond} //sendReportDataTimeoutSecond秒超时
    resp, errDo := cli.Do(request)
    if err != nil || nil == resp {
        logrus.WithFields(logrus.Fields{"err": errDo, "url": url, "body": body}).Error("http client Do failed")
        return errDo
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        logrus.WithFields(logrus.Fields{"status": resp.StatusCode, "url": url, "body": body}).Error("http resp status error")
        return errors.New("http response status error")
    }
    rspBody, errRead := ioutil.ReadAll(resp.Body)
    if errRead != nil {
        logrus.WithFields(logrus.Fields{"err": errRead, "url": url, "body": body}).Error("read http response failed")
    }
    logrus.WithFields(logrus.Fields{"rspBody": string(rspBody), "url": url, "reqBody": body}).Debug("http request success")
    return nil
}

func getSuccCountAliyunMetrics(succCountMap map[string]countInfo, reportTime time.Time) []AliyunMetric {
    metrics := make([]AliyunMetric, 0)
    for key, countObj := range succCountMap {
        dimessionObj := parseSuccessCountDimension(key)
        if nil == dimessionObj {
            logrus.WithFields(logrus.Fields{"key": key}).Warn("report success count dimession key abnormal")
            continue
        }
        if "" != dimessionObj.SName {
            dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
        }
        if "" != dimessionObj.TName {
            dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
        }

        var metric AliyunMetric
        metric.MetricName = dimessionObj.getMetricName()
        metric.Value = countObj.counter
        metric.Timestamp = reportTime.UnixNano() / 1e6
        metric.Dimensions = dimessionObj
        metrics = append(metrics, metric)
    }
    return metrics
}

func getFailedCountAliyunMetrics(failedCountMap map[string]countInfo, reportTime time.Time) []AliyunMetric {
    metrics := make([]AliyunMetric, 0)
    for key, countObj := range failedCountMap {
        dimessionObj := parseFailedCountDimension(key)
        if nil == dimessionObj {
            logrus.WithFields(logrus.Fields{"key": key}).Warn("report failed count dimession key abnormal")
            continue
        }
        if "" != dimessionObj.SName {
            dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
        }
        if "" != dimessionObj.TName {
            dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
        }

        var metric AliyunMetric
        metric.MetricName = dimessionObj.getMetricName()
        metric.Value = countObj.counter
        metric.Timestamp = reportTime.UnixNano() / 1e6
        metric.Dimensions = dimessionObj
        metrics = append(metrics, metric)
    }
    return metrics
}

func getSuccAvgTimeAliyunMetrics(succAvgTimeMap map[string]countInfo, reportTime time.Time) []AliyunMetric {
    metrics := make([]AliyunMetric, 0)
    for key, countObj := range succAvgTimeMap {
        dimessionObj := parseSuccessAvgTimeDimension(key)
        if nil == dimessionObj || countObj.counter <= 0 {
            logrus.WithFields(logrus.Fields{"key": key}).Warn("report success avg time dimession abnormal")
            continue
        }
        if "" != dimessionObj.SName {
            dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
        }
        if "" != dimessionObj.TName {
            dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
        }

        var metric AliyunMetric
        metric.MetricName = dimessionObj.getMetricName()
        metric.Value = countObj.sum / countObj.counter //平均值
        metric.Timestamp = reportTime.UnixNano() / 1e6
        metric.Dimensions = dimessionObj
        metrics = append(metrics, metric)
    }
    return metrics
}

func getFailedAvgTimeAliyunMetrics(failedAvgTimeMap map[string]countInfo, reportTime time.Time) []AliyunMetric {
    metrics := make([]AliyunMetric, 0)
    for key, countObj := range failedAvgTimeMap {
        dimessionObj := parseFailedAvgTimeDimension(key)
        if nil == dimessionObj || countObj.counter <= 0 {
            logrus.WithFields(logrus.Fields{"key": key}).Warn("report failed avg time dimession abnormal")
            continue
        }
        if "" != dimessionObj.SName {
            dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
        }
        if "" != dimessionObj.TName {
            dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
        }

        var metric AliyunMetric
        metric.MetricName = dimessionObj.getMetricName()
        metric.Value = countObj.sum / countObj.counter //平均值
        metric.Timestamp = reportTime.UnixNano() / 1e6
        metric.Dimensions = dimessionObj
        metrics = append(metrics, metric)
    }
    return metrics
}

func sendMetricsToStatsd(metrics []string) error {
    if len(metrics) <= 0 {
        return nil
    }
    conn, err := net.Dial("udp", monitorObj.conf.StatsdAddr)
    if err != nil {
        logrus.WithFields(logrus.Fields{"StatsdAddr": monitorObj.conf.StatsdAddr, "error": err}).Error("conn statsd failed")
        return err
    }

    for index, _ := range metrics {
        _, err := fmt.Fprintf(conn, metrics[index])
        if err != nil {
            logrus.WithFields(logrus.Fields{"content": metrics[index], "error": err}).Error("Send statsd metric failed")
        }
    }
    logrus.WithFields(logrus.Fields{"metrics": metrics}).Debug("Send statsd metrics complete")
    return nil
}

func getSuccCountStatsdMetrics(succCountMap map[string]countInfo, reportTime time.Time) []string {
    metrics := make([]string, 0)
    for key, countObj := range succCountMap {
        key = strings.Replace(key, ".", "-", -1) //上报到statsd，label中不能有"."
        dimessionObj := parseSuccessCountDimension(key)
        if nil == dimessionObj {
            logrus.WithFields(logrus.Fields{"key": key}).Warn("report success count dimession key abnormal")
            continue
        }
        if "" != dimessionObj.SName {
            dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
        }
        if "" != dimessionObj.TName {
            dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
        }

        metric := fmt.Sprintf("req.success.count.%v.%v.%v.%v.%v:%v|c", dimessionObj.SName, dimessionObj.SIP, dimessionObj.TName, dimessionObj.TIP, strings.Replace(dimessionObj.Infc,":","-", -1), countObj.counter)
        metrics = append(metrics, metric)
    }
    return metrics
}

func getFailedCountStatsdMetrics(failedCountMap map[string]countInfo, reportTime time.Time) []string {
    metrics := make([]string, 0)
    for key, countObj := range failedCountMap {
        key = strings.Replace(key, ".", "-", -1) //上报到statsd，label中不能有"."
        dimessionObj := parseFailedCountDimension(key)
        if nil == dimessionObj {
            logrus.WithFields(logrus.Fields{"key": key}).Warn("report failed count dimession key abnormal")
            continue
        }
        if "" != dimessionObj.SName {
            dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
        }
        if "" != dimessionObj.TName {
            dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
        }

        metric := fmt.Sprintf("req.failed.count.%v.%v.%v.%v.%v:%v|c", dimessionObj.SName, dimessionObj.TName, dimessionObj.TIP, dimessionObj.Code, strings.Replace(dimessionObj.Infc, ":", "-", -1), countObj.counter)
        metrics = append(metrics, metric)
    }
    return metrics
}

func getSuccAvgTimeStatsdMetrics(succAvgTimeMap map[string]countInfo, reportTime time.Time) []string {
    metrics := make([]string, 0)
    for key, countObj := range succAvgTimeMap {
        key = strings.Replace(key, ".", "-", -1) //上报到statsd，label中不能有"."
        dimessionObj := parseSuccessAvgTimeDimension(key)
        if nil == dimessionObj || countObj.counter <= 0 {
            logrus.WithFields(logrus.Fields{"key": key}).Warn("report success avg time dimession abnormal")
            continue
        }
        if "" != dimessionObj.SName {
            dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
        }
        if "" != dimessionObj.TName {
            dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
        }

        metric := fmt.Sprintf("req.success.avg.time.%v.%v.%v.%v.%v:%v|ms", dimessionObj.SName, dimessionObj.SIP, dimessionObj.TName, dimessionObj.TIP, strings.Replace(dimessionObj.Infc, ":", "-", -1), countObj.sum/countObj.counter)
        metrics = append(metrics, metric)
    }
    return metrics
}

func getFailedAvgTimeStatsdMetrics(failedAvgTimeMap map[string]countInfo, reportTime time.Time) []string {
    metrics := make([]string, 0)
    for key, countObj := range failedAvgTimeMap {
        key = strings.Replace(key, ".", "-", -1) //上报到statsd，label中不能有"."
        dimessionObj := parseFailedAvgTimeDimension(key)
        if nil == dimessionObj || countObj.counter <= 0 {
            logrus.WithFields(logrus.Fields{"key": key}).Warn("report failed avg time dimession abnormal")
            continue
        }
        if "" != dimessionObj.SName {
            dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
        }
        if "" != dimessionObj.TName {
            dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
        }

        metric := fmt.Sprintf("req.failed.avg.time.%v.%v.%v.%v.%v:%v|ms", dimessionObj.SName, dimessionObj.SIP, dimessionObj.TName, dimessionObj.TIP, strings.Replace(dimessionObj.Infc, ":", "-", -1), countObj.sum/countObj.counter)
        metrics = append(metrics, metric)
    }
    return metrics
}
