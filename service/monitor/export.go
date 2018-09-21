package monitor

import (
    "errors"
    "net"
    "time"
    "strings"
)

// MonitorConf 监控配置
type MonitorConf struct {
    EnableReport    bool   // 启用上报到阿里云监控
    CurServiceName  string // 本服务的名称
    CurServerIP     string // 本机IP
    EnvironmentType string // 环境类型,test/qa/prod
    AliUid          string // 上报到阿里云监控的Uid
    AliNamespace    string // 上报到阿里云监控的namespace
    ReportAddr      string // 上报地址，默认："open.cms.aliyun.com"
    EnableStatsd    bool   // 是否启用上报到statsd
    StatsdAddr      string // statsd上报地址，默认："metrics.lwork.com:9125"
}

// ReqSuccessCountDimension 请求成功量的监控项字段
type ReqSuccessCountDimension struct {
    SName string `json:"sName,omitempty"` // 源端服务名称，若非空则上报到阿里云时：SName = "{SName}_{environmen_type}"
    SIP   string `json:"sIP,omitempty"`   // 源端IP地址
    TName string `json:"tName,omitempty"` // 目标端服务名称，若非空则上报到阿里云时：TName = "{TName}_{environmen_type}"
    TIP   string `json:"tIP,omitempty"`   // 目标端IP地址
    Infc  string `json:"infc,omitempty"`  // {direction}_{method}_{path}"，method取值：direction：ACTIVE/PASSIVE，GET/POST/DELETE/PUT/RPC/GRPC， path：Http Path/Interface Name；例如：ACTIVE_GET_/v1/user/right
}

// ReqFailedCountDimension 请求失败量的监控项字段
type ReqFailedCountDimension struct {
    SName string `json:"sName,omitempty"` // 源端服务名称，若非空则上报到阿里云时：SName = "{SName}_{environmen_type}"
    TName string `json:"tName,omitempty"` // 目标端服务名称，若非空则上报到阿里云时：TName = "{TName}_{environmen_type}"
    TIP   string `json:"tIP,omitempty"`   // 目标端IP地址
    Code  string `json:"code,omitempty"`  // 错误码
    Infc  string `json:"infc,omitempty"`  // {direction}_{method}_{path}"，method取值：direction：ACTIVE/PASSIVE，GET/POST/DELETE/PUT/RPC/GRPC， path：Http Path/Interface Name；例如：ACTIVE_GET_/v1/user/right
}

// ReqSuccessAvgTimeDimension 请求成功平均耗时的监控项字段
type ReqSuccessAvgTimeDimension struct {
    SName string `json:"sName,omitempty"` // 源端服务名称，若非空则上报到阿里云时：SName = "{SName}_{environmen_type}"
    SIP   string `json:"sIP,omitempty"`   // 源端IP地址，若非空则上报到阿里云时：TName = "{TName}_{environmen_type}"
    TName string `json:"tName,omitempty"` // 目标端服务名称
    TIP   string `json:"tIP,omitempty"`   // 目标端IP地址
    Infc  string `json:"infc,omitempty"`  // {direction}_{method}_{path}"，method取值：direction：ACTIVE/PASSIVE，GET/POST/DELETE/PUT/RPC/GRPC， path：Http Path/Interface Name；例如：ACTIVE_GET_/v1/user/right
}

// ReqFailedAvgTimeDimension 请求失败平均耗时的监控项字段
type ReqFailedAvgTimeDimension struct {
    SName string `json:"sName,omitempty"` // 源端服务名称，若非空则上报到阿里云时：SName = "{SName}_{environmen_type}"
    SIP   string `json:"sIP,omitempty"`   // 源端IP地址
    TName string `json:"tName,omitempty"` // 目标端服务名称，若非空则上报到阿里云时：TName = "{TName}_{environmen_type}"
    TIP   string `json:"tIP,omitempty"`   // 目标端IP地址
    Infc  string `json:"infc,omitempty"`  // {direction}_{method}_{path}"，method取值：direction：ACTIVE/PASSIVE，GET/POST/DELETE/PUT/RPC/GRPC， path：Http Path/Interface Name；例如：ACTIVE_GET_/v1/user/right
}

type RuntimeDataDimension struct {
    SerName string   `json:"serName,omitempty"`
    SerIP   string   `json:"serIP,omitempty"`
    Option  string   `json:"option,omitempty"`
}

// Init 监控初始化
func Init(conf *MonitorConf) error {
    if nil == conf || (false == conf.EnableReport && false == conf.EnableStatsd) {
        return nil
    }
    if true == conf.EnableReport && ("" == conf.EnvironmentType || "" == conf.CurServiceName || "" == conf.AliUid || "" == conf.AliNamespace) {
        return errors.New("Monitor Init Parameter Error")
    }
    if true == conf.EnableStatsd && ("" == conf.EnvironmentType || "" == conf.CurServiceName) {
        return errors.New("Monitor Init Parameter Error")
    }
    monitorObj.conf = *conf

    if "" == conf.ReportAddr {
        monitorObj.conf.ReportAddr = defaultAliReportAddr
    }
    if strings.HasSuffix(monitorObj.conf.ReportAddr, "/") {
        monitorObj.conf.ReportAddr = monitorObj.conf.ReportAddr[:len(monitorObj.conf.ReportAddr)-1] //去除最后一个"/"字符
    }

    if "" == conf.StatsdAddr {
        monitorObj.conf.StatsdAddr = defaultStatsdReportAddr
    }
    if strings.HasSuffix(monitorObj.conf.StatsdAddr, "/") {
        monitorObj.conf.StatsdAddr = monitorObj.conf.StatsdAddr[:len(monitorObj.conf.StatsdAddr)-1] //去除最后一个"/"字符
    }

    monitorObj.reqSuccCountChan = make(chan *ReqSuccessCountDimension, reportQueueLength)
    monitorObj.reqFailedCountChan = make(chan *ReqFailedCountDimension, reportQueueLength)
    monitorObj.reqSuccTimeConsumeChan = make(chan *reqSuccessTimeConsumeInfo, reportQueueLength)
    monitorObj.reqFailedTimeConsumeChan = make(chan *reqFailedTimeConsumeInfo, reportQueueLength)
    monitorObj.succCountMap = make(map[string]countInfo)
    monitorObj.failedCountMap = make(map[string]countInfo)
    monitorObj.succAvgTimeMap = make(map[string]countInfo)
    monitorObj.failedAvgTimeMap = make(map[string]countInfo)
    go monitorObj.processReportData() // 启动一个协程处理上报数据

    return nil
}

// GetCurrentServerName 返回服务名称
func GetCurrentServerName() string {
    return monitorObj.conf.CurServiceName
}

// GetCurrentServerIP 返回服务器IP
func GetCurrentServerIP() string {
    return monitorObj.conf.CurServerIP
}

// EnableReportMonitor 返回监控是否上报
func EnableReportMonitor() bool {
    return monitorObj.conf.EnableReport || monitorObj.conf.EnableStatsd
}

// IsInnerIPv4 判断一个IPv4地址是否为内网
func IsInnerIPv4(IPv4Str string) bool {
    ipv4 := net.ParseIP(IPv4Str).To4()
    if nil == ipv4 || len(ipv4) != 4 {
        return false
    }
    var ipNum uint32
    ipNum = (uint32(ipv4[0]) << 24) + (uint32(ipv4[1]) << 16) + (uint32(ipv4[2]) << 8) + uint32(ipv4[3])
    //167772160:10.0.0.0, 184549375:10.255.255.255, 3232235520:192.168.0.0, 3232301055:192.168.255.255, 2886729728:172.16.0.0, 2887778303:172.31.255.255, 2130706433:127.0.0.1
    if (ipNum >= 167772160 && ipNum <= 184549375) || (ipNum >= 3232235520 && ipNum <= 3232301055) || (ipNum >= 2886729728 && ipNum <= 2887778303) || (2130706433 == ipNum) {
        return true
    }
    return false
}

// ReportReqSuccess 上报请求成功
func ReportReqSuccess(oneData *ReqSuccessCountDimension) {
    if nil == oneData || false == EnableReportMonitor() {
        return
    }
    oneData.SIP = ""  //为减少序列数，暂时停止IP上报
    oneData.TIP = ""
    select {
    case monitorObj.reqSuccCountChan <- oneData:
        //do nothing
    default:
        //do nothing, 防止reqSuccCountChan已满而阻塞
    }
}

// ReportReqFailed 上报请求失败
func ReportReqFailed(oneData *ReqFailedCountDimension) {
    if nil == oneData || false == EnableReportMonitor() {
        return
    }
    oneData.TIP = ""//为减少序列数，暂时停止IP上报
    select {
    case monitorObj.reqFailedCountChan <- oneData:
        //do nothing
    default:
        //do nothing, 防止reqFailedCountChan已满而阻塞
    }
}

// ReportSuccessAvgTime 上报请求成功耗时
// timeConsume：耗时（微秒）
func ReportSuccessAvgTime(oneData *ReqSuccessAvgTimeDimension, timeConsume int64) {
    if nil == oneData || false == EnableReportMonitor() || timeConsume <= 0 {
        return
    }
    oneData.SIP = ""  //为减少序列数，暂时停止IP上报
    oneData.TIP = ""
    var temp reqSuccessTimeConsumeInfo
    temp.succAvgTimeDimension = oneData
    temp.timeConsume = timeConsume
    select {
    case monitorObj.reqSuccTimeConsumeChan <- &temp:
        //do nothing
    default:
        //do nothing, 防止reqSuccTimeConsumeChan已满而阻塞
    }
}

// ReportFailedAvgTime 上报请求失败耗时
// timeConsume：耗时（微秒）
func ReportFailedAvgTime(oneData *ReqFailedAvgTimeDimension, timeConsume int64) {
    if nil == oneData || false == EnableReportMonitor() {
        return
    }
    oneData.SIP = ""  //为减少序列数，暂时停止IP上报
    oneData.TIP = ""
    var temp reqFailedTimeConsumeInfo
    temp.failedAvgTimeDimension = oneData
    temp.timeConsume = timeConsume
    select {
    case monitorObj.reqFailedTimeConsumeChan <- &temp:
        //do nothing
    default:
        //do nothing, 防止reqFailedTimeConsumeChan已满而阻塞
    }
}

// CommMonitorReport 成功上报接口，会上报请求成功计数和成功平均耗时
func CommMonitorReport(errCode, sName, sIP, tName, tIP, infc string, beginTime time.Time) error {
    if false == EnableReportMonitor() {
        return nil
    }
    if ("" == sName && "" == sIP && "" == tName && "" == tIP && "" == infc) || (0 == beginTime.UnixNano()) {
        return errors.New("Parameter error")
    }
    if "" == errCode {
        var succCountReport ReqSuccessCountDimension
        succCountReport.SName = sName
        succCountReport.SIP = sIP
        succCountReport.TName = tName
        succCountReport.TIP = tIP
        succCountReport.Infc = infc
        ReportReqSuccess(&succCountReport)

        var succAvgTimeReport ReqSuccessAvgTimeDimension
        succAvgTimeReport.SName = sName
        succAvgTimeReport.SIP = sIP
        succAvgTimeReport.TName = tName
        succAvgTimeReport.TIP = tIP
        succAvgTimeReport.Infc = infc
        ReportSuccessAvgTime(&succAvgTimeReport, (time.Now().UnixNano()-beginTime.UnixNano())/1e3) //耗时单位为微秒
    } else { // "" != errCode
        var failedCountReport ReqFailedCountDimension
        failedCountReport.SName = sName
        failedCountReport.TName = tName
        failedCountReport.TIP = tIP
        failedCountReport.Code = errCode
        failedCountReport.Infc = infc
        ReportReqFailed(&failedCountReport)

        var failedAvgTimeReport ReqFailedAvgTimeDimension
        failedAvgTimeReport.SName = sName
        failedAvgTimeReport.SIP = sIP
        failedAvgTimeReport.TName = tName
        failedAvgTimeReport.TIP = tIP
        failedAvgTimeReport.Infc = infc
        ReportFailedAvgTime(&failedAvgTimeReport, (time.Now().UnixNano()-beginTime.UnixNano())/1e3) //耗时单位为微秒
    }

    return nil
}
