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
)

const (
	reportInterval                = 3    // 上报到阿里云的时间间隔，3秒
	checkReportDataCountLimit     = 1000 // 当上报checkReportDataCountLimit记录后，检查一次是否该发送数据
	notReportDataSleepMillisecond = 10   // 没有上报数据时的休眠时间，单位毫秒，避免循环消耗太多CPU
	delimit                       = "#@#"
	defaultReportAddr             = "open.cms.aliyun.com"
	reportQueneLength             = 300 // 上报数据队列的长度
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
	monitorObj           monitorInfo
	lastSendToAliyunTime time.Time
)

// checkAndReportDataToAliyun 检查是否需要发送上报数据到阿里云，若需要则发送后返回并且修改lastSendToAliyunTime为当前时间
func (me *monitorInfo) checkAndReportDataToAliyun() bool {
	timeNow := time.Now()
	if timeNow.Unix()-lastSendToAliyunTime.Unix() < reportInterval {
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
	go reportSuccCountToAliyun(succCountMap, timeNow)
	go reportFailedCountToAliyun(failedCountMap, timeNow)
	go reportSuccAvgTimeToAliyun(succAvgTimeMap, timeNow)
	go reportFailedAvgTimeToAliyun(failedAvgTimeMap, timeNow)
	lastSendToAliyunTime = time.Now()

	return true
}

// processReportData 处理上报数据的函数
func (me *monitorInfo) processReportData() {
	reportCountTmp := 0
	lastSendToAliyunTime = time.Now()

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
			if reportCountTmp > checkReportDataCountLimit && me.checkAndReportDataToAliyun() { // 避免上报数据太多，长时间没机会执行reportDataToAliyun
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
			if reportCountTmp > checkReportDataCountLimit && me.checkAndReportDataToAliyun() { // 避免上报数据太多，长时间没机会执行reportDataToAliyun
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
			if reportCountTmp > checkReportDataCountLimit && me.checkAndReportDataToAliyun() { // 避免上报数据太多，长时间没机会执行reportDataToAliyun
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
			if reportCountTmp > checkReportDataCountLimit && me.checkAndReportDataToAliyun() { // 避免上报数据太多，长时间没机会执行reportDataToAliyun
				reportCountTmp = 0
			}
		default:
			if me.checkAndReportDataToAliyun() { // 避免上报数据太多，长时间没机会执行reportDataToAliyun
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

func sendRequestToAliMonitor(metrics []AliyunMetric) error {
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

func reportSuccCountToAliyun(succCountMap map[string]countInfo, reportTime time.Time) {
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
	if len(metrics) > 0 {
		sendRequestToAliMonitor(metrics)
	}
}

func reportFailedCountToAliyun(failedCountMap map[string]countInfo, reportTime time.Time) {
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
	if len(metrics) > 0 {
		sendRequestToAliMonitor(metrics)
	}
}

func reportSuccAvgTimeToAliyun(succAvgTimeMap map[string]countInfo, reportTime time.Time) {
	metrics := make([]AliyunMetric, 0)
	for key, countObj := range succAvgTimeMap {
		dimessionObj := parseSuccessAvgTimeDimension(key)
		if nil == dimessionObj {
			logrus.WithFields(logrus.Fields{"key": key}).Warn("report success avg time dimession key abnormal")
			continue
		}
		if "" != dimessionObj.SName {
			dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
		}
		if "" != dimessionObj.TName {
			dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
		}
		if countObj.counter <= 0 {
			logrus.WithFields(logrus.Fields{"counter": countObj.counter}).Warn("report success avg time counter abnormal")
			continue
		}

		var metric AliyunMetric
		metric.MetricName = dimessionObj.getMetricName()
		metric.Value = countObj.sum / countObj.counter //平均值
		metric.Timestamp = reportTime.UnixNano() / 1e6
		metric.Dimensions = dimessionObj
		metrics = append(metrics, metric)
	}
	if len(metrics) > 0 {
		sendRequestToAliMonitor(metrics)
	}
}

func reportFailedAvgTimeToAliyun(failedAvgTimeMap map[string]countInfo, reportTime time.Time) {
	metrics := make([]AliyunMetric, 0)
	for key, countObj := range failedAvgTimeMap {
		dimessionObj := parseFailedAvgTimeDimension(key)
		if nil == dimessionObj {
			logrus.WithFields(logrus.Fields{"key": key}).Warn("report failed avg time dimession key abnormal")
			continue
		}
		if "" != dimessionObj.SName {
			dimessionObj.SName += "_" + monitorObj.conf.EnvironmentType
		}
		if "" != dimessionObj.TName {
			dimessionObj.TName += "_" + monitorObj.conf.EnvironmentType
		}
		if countObj.counter <= 0 {
			logrus.WithFields(logrus.Fields{"counter": countObj.counter}).Warn("report failed avg time counter abnormal")
			continue
		}

		var metric AliyunMetric
		metric.MetricName = dimessionObj.getMetricName()
		metric.Value = countObj.sum / countObj.counter //平均值
		metric.Timestamp = reportTime.UnixNano() / 1e6
		metric.Dimensions = dimessionObj
		metrics = append(metrics, metric)
	}
	if len(metrics) > 0 {
		sendRequestToAliMonitor(metrics)
	}
}
