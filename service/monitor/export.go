package monitor

import (
	"errors"
	"net"
	"time"
)


type MoniorConf struct {
	EnableReport   bool     		//启用上报到阿里云监控
	CurServiceName string           //本服务的名称
	CurServerIP    string           //本机IP
	EnvironmenType string			//环境类型,dev/qa/online
	AliUid         string           //上报到阿里云监控的Uid
	AliNamespace   string          //上报到阿里云监控的namespace
}


//请求成功量的监控项字段
type ReqSuccessCountDimension struct {
	SName      string      `json:"sName,omitempty"` //源端服务名称，若非空则上报到阿里云时：SName = "{SName}_{environmen_type}"
	SIP        string      `json:"sIP,omitempty"`   //源端IP地址
	TName      string      `json:"tName,omitempty"` //目标端服务名称，若非空则上报到阿里云时：TName = "{TName}_{environmen_type}"
	TIP        string      `json:"tIP,omitempty"`   //目标端IP地址
	Infc       string      `json:"infc,omitempty"`  //{direction}_{method}_{path}"，method取值：direction：ACTIVE/PASSIVE，GET/POST/DELETE/PUT/RPC/GRPC， path：Http Path/Interface Name；例如：ACTIVE_GET_/v1/user/right
}
//请求失败量的监控项字段
type ReqFailedCountDimension struct {
	SName      string      `json:"sName,omitempty"` //源端服务名称，若非空则上报到阿里云时：SName = "{SName}_{environmen_type}"
	TName      string      `json:"tName,omitempty"` //目标端服务名称，若非空则上报到阿里云时：TName = "{TName}_{environmen_type}"
	TIP        string      `json:"tIP,omitempty"`   //目标端IP地址
	Code       string      `json:"code,omitempty"`  //错误码
	Infc       string      `json:"infc,omitempty"`  //{direction}_{method}_{path}"，method取值：direction：ACTIVE/PASSIVE，GET/POST/DELETE/PUT/RPC/GRPC， path：Http Path/Interface Name；例如：ACTIVE_GET_/v1/user/right
}
//请求成功平均耗时的监控项字段
type ReqSuccessAvgTimeDimension struct {
	SName      string      `json:"sName,omitempty"` //源端服务名称，若非空则上报到阿里云时：SName = "{SName}_{environmen_type}"
	SIP        string      `json:"sIP,omitempty"`   //源端IP地址，若非空则上报到阿里云时：TName = "{TName}_{environmen_type}"
	TName      string      `json:"tName,omitempty"` //目标端服务名称
	TIP        string      `json:"tIP,omitempty"`   //目标端IP地址
	Infc       string      `json:"infc,omitempty"`  //{direction}_{method}_{path}"，method取值：direction：ACTIVE/PASSIVE，GET/POST/DELETE/PUT/RPC/GRPC， path：Http Path/Interface Name；例如：ACTIVE_GET_/v1/user/right
}
//请求失败平均耗时的监控项字段
type ReqFailedAvgTimeDimension struct {
	SName      string      `json:"sName,omitempty"` //源端服务名称，若非空则上报到阿里云时：SName = "{SName}_{environmen_type}"
	SIP        string      `json:"sIP,omitempty"`   //源端IP地址
	TName      string      `json:"tName,omitempty"` //目标端服务名称，若非空则上报到阿里云时：TName = "{TName}_{environmen_type}"
	TIP        string      `json:"tIP,omitempty"`   //目标端IP地址
	Infc       string      `json:"infc,omitempty"`  //{direction}_{method}_{path}"，method取值：direction：ACTIVE/PASSIVE，GET/POST/DELETE/PUT/RPC/GRPC， path：Http Path/Interface Name；例如：ACTIVE_GET_/v1/user/right
}


//monitor监控初始化
func Init(conf *MoniorConf) error {
	if nil == conf || (true == conf.EnableReport && ("" == conf.EnvironmenType || "" == conf.CurServiceName || "" == conf.AliUid || "" == conf.AliNamespace)) {
		return errors.New("Monitor Init Parameter Error")
	}
	monitorObj.conf = *conf
	if false == monitorObj.conf.EnableReport {
		return nil
	}

	monitorObj.reqSuccCountChan = make(chan *ReqSuccessCountDimension, reportQueneLength)
	monitorObj.reqFailedCountChan = make(chan *ReqFailedCountDimension, reportQueneLength)
	monitorObj.reqSuccTimeConsumeChan = make(chan *reqSuccessTimeConsumeInfo, reportQueneLength)
	monitorObj.reqFailedTimeConsumeChan = make(chan *reqFailedTimeConsumeInfo, reportQueneLength)
	monitorObj.succCountMap = make(map[string]countInfo)
	monitorObj.failedCountMap = make(map[string]countInfo)
	monitorObj.succAvgTimeMap = make(map[string]countInfo)
	monitorObj.failedAvgTimeMap = make(map[string]countInfo)
	go monitorObj.processReportData()		//启动一个协程处理上报数据
	return nil
}

func GetCurrentServerName() string {
	return monitorObj.conf.CurServiceName
}
func GetCurrentServerIP() string {
	return monitorObj.conf.CurServerIP
}
func EnableReportMonitor() bool {
	return monitorObj.conf.EnableReport
}
//判断一个IPv4地址是否为内网
func IsInnerIPv4(IPv4Str string) bool {
	ipv4 := net.ParseIP(IPv4Str).To4()
	if nil == ipv4 || len(ipv4) != 4 {
		return false
	}
	var ipNum uint32
	ipNum = (uint32(ipv4[0])<<24) + (uint32(ipv4[1])<<16) + (uint32(ipv4[2])<<8) + uint32(ipv4[3])
	//167772160:10.0.0.0, 184549375:10.255.255.255, 3232235520:192.168.0.0, 3232301055:192.168.255.255, 2886729728:172.16.0.0, 2887778303:172.31.255.255, 2130706433:127.0.0.1
	if (ipNum >= 167772160 && ipNum <= 184549375) || (ipNum >= 3232235520 && ipNum <= 3232301055) || (ipNum >= 2886729728 && ipNum <= 2887778303) || (2130706433 == ipNum) {
		return  true
	}
	return false
}

func ReportReqSuccess(oneData *ReqSuccessCountDimension) {
	if nil == oneData || false == monitorObj.conf.EnableReport {
		return
	}
	select {
	case monitorObj.reqSuccCountChan<-oneData:
		//do nothing
	default:
		//do nothing, 防止reqSuccCountChan已满而阻塞
	}
}

func ReportReqFailed(oneData *ReqFailedCountDimension) {
	if nil == oneData || false == monitorObj.conf.EnableReport {
		return
	}
	select {
	case monitorObj.reqFailedCountChan<-oneData:
		//do nothing
	default:
		//do nothing, 防止reqFailedCountChan已满而阻塞
	}
}

//timeConsume：耗时（微秒）
func ReportSuccessAvgTime(oneData *ReqSuccessAvgTimeDimension, timeConsume int64) {
	if nil == oneData || false == monitorObj.conf.EnableReport || timeConsume <= 0 {
		return
	}
	var temp reqSuccessTimeConsumeInfo
	temp.succAvgTimeDimension = oneData
	temp.timeConsume = timeConsume
	select {
	case monitorObj.reqSuccTimeConsumeChan<-&temp:
		//do nothing
	default:
		//do nothing, 防止reqSuccTimeConsumeChan已满而阻塞
	}
}


//timeConsume：耗时（微秒）
func ReportFailedAvgTime(oneData *ReqFailedAvgTimeDimension, timeConsume int64) {
	if nil == oneData || false == monitorObj.conf.EnableReport {
		return
	}
	var temp reqFailedTimeConsumeInfo
	temp.failedAvgTimeDimension = oneData
	temp.timeConsume = timeConsume
	select {
	case monitorObj.reqFailedTimeConsumeChan<-&temp:
		//do nothing
	default:
		//do nothing, 防止reqFailedTimeConsumeChan已满而阻塞
	}
}



//成功上报接口，会上报请求成功计数和成功平均耗时
func CommMonitorReport(errCode, sName, sIP, tName, tIP, infc string, beginTime time.Time) error {
	if false == monitorObj.conf.EnableReport {
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
		ReportSuccessAvgTime(&succAvgTimeReport, (time.Now().UnixNano() - beginTime.UnixNano()) / 1e3)		//耗时单位为微秒
	} else {			// "" != errCode
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
		ReportFailedAvgTime(&failedAvgTimeReport, (time.Now().UnixNano() - beginTime.UnixNano()) / 1e3) //耗时单位为微秒
	}

	return nil
}