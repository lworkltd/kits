package monitor

import (
	"errors"
	"strings"
	"sync"
	"time"
	"github.com/Sirupsen/logrus"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"fmt"
)

const (
	reportInterval = 1	//上报到阿里云的时间间隔，2秒
	delimit = "#@#"
	uid ="1765747156115092"
	namespace ="ACS/CUSTOM/1765747156115092"
	endpoint ="open.cms.aliyun.com"
)

//简易序列化
func (this *ReqSuccessCountDimension) generatekey() string {
	return this.SName + delimit + this.SIP + delimit + this.TName + delimit + this.TIP + delimit + this.Infc
}
//简易反序列化
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
func (this *ReqSuccessCountDimension) getMetricName() string {
	return "req_success_count"
}


//简易序列化
func (this *ReqFailedCountDimension) generatekey() string {
	return this.SName + delimit + this.TName + delimit + this.TIP + delimit + this.Code + delimit + this.Infc
}
//简易反序列化
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
func (this *ReqFailedCountDimension) getMetricName() string {
	return "req_failed_count"
}


//简易序列化
func (this *ReqSuccessAvgTimeDimension) generatekey() string {
	return this.SName + delimit + this.SIP + delimit + this.TName + delimit + this.TIP + delimit + this.Infc
}
//简易反序列化
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
func (this *ReqSuccessAvgTimeDimension) getMetricName() string {
	return "req_success_avg_time"
}


//简易序列化
func (this *ReqFailedAvgTimeDimension) generatekey() string {
	return this.SName + delimit + this.SIP + delimit + this.TName + delimit + this.TIP + delimit + this.Infc
}
//简易反序列化
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
func (this *ReqFailedAvgTimeDimension) getMetricName() string {
	return "req_failed_avg_time"
}

//用于计数
type countInfo struct {
	counter       int64			//次数
	sum           int64			//总和，例如耗时总和（微秒）
}

type monitorInfo struct {
	conf             MoniorConf
	succCountMap     map[string]countInfo		//请求成功次数计数，key为ReqSuccessCountDimension序列化字符串
	failedCountMap   map[string]countInfo		//请求失败次数计数，key为ReqFailedCountDimension序列化后字符串
	succAvgTimeMap   map[string]countInfo		//请求成功平均耗时计数，key为ReqSuccessAvgTimeDimension序列化字符串
	failedAvgTimeMap map[string]countInfo		//请求失败平均耗时计数，key为ReqFailedAvgTimeDimension序列化字符串
	mutex            sync.RWMutex				//操作map计数的锁
}

var (
	monitorObj       monitorInfo
)



func (this *monitorInfo)reportDataToAliyun() {
	for {
		time.Sleep(time.Second * reportInterval)
		{
			this.mutex.Lock()			//加锁，从缓存取出已记录的待上报数据去上报，取出后清理缓存重新计数
			succCountMap := this.succCountMap
			failedCountMap := this.failedCountMap
			succAvgTimeMap := this.succAvgTimeMap
			failedAvgTimeMap := this.failedAvgTimeMap
			this.succCountMap = make(map[string]countInfo)
			this.failedCountMap = make(map[string]countInfo)
			this.succAvgTimeMap = make(map[string]countInfo)
			this.failedAvgTimeMap = make(map[string]countInfo)
			this.mutex.Unlock()
			timeNow := time.Now()
			go reportSuccCountToAliyun(succCountMap, timeNow)
			go reportFailedCountToAliyun(failedCountMap, timeNow)
			go reportSuccAvgTimeToAliyun(succAvgTimeMap, timeNow)
			go reportFailedAvgTimeToAliyun(failedAvgTimeMap, timeNow)
		}
	}
}


type AliyunMetric struct {
	MetricName    string           `json:"metricName"`
	Value         int64            `json:"value"`
	Timestamp     int64            `json:"timestamp"`
	Unit          string           `json:"unit"`
	Dimensions    interface{}     `json:"dimensions"`
}
type AliyunMonitorInfo struct{
	UserId        string           `json:"userId"`
	Namespace     string           `json:"namespace"`
	Metrics       []AliyunMetric   `json:"metrics"`
}


func sendRequestToAliMonitor(bodyObj *AliyunMonitorInfo) error {
	metricsBytes, _ := json.Marshal(bodyObj.Metrics)
	body := fmt.Sprintf("userId=%v&namespace=%v&metrics=%v", bodyObj.UserId, bodyObj.Namespace, string(metricsBytes))
	url := "http://" + endpoint + "/metrics/put"
	request, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		logrus.WithFields(logrus.Fields{"err":err, "url":url, "body":body}).Error("http.NewRequest failed")
		return err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Connection", "close")

	cli := &http.Client{Timeout:time.Second * 2}			//2秒超时
	resp, errDo := cli.Do(request)
	if err != nil || nil == resp {
		logrus.WithFields(logrus.Fields{"err":errDo, "url":url, "body":body}).Error("http client Do failed")
		return errDo
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{"status":resp.StatusCode, "url":url, "body":body}).Error("http resp status error")
		return errors.New("http response status error")
	}
	rspBody, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		logrus.WithFields(logrus.Fields{"err":errRead, "url":url, "body":body}).Error("read http response failed")
	}
	logrus.WithFields(logrus.Fields{"rspBody":string(rspBody), "url":url, "reqBody":body}).Debug("http request success")
	return nil
}

func reportSuccCountToAliyun(succCountMap map[string]countInfo, reportTime time.Time) {
	var reportInfo AliyunMonitorInfo
	reportInfo.Namespace = namespace
	reportInfo.UserId = uid
	reportInfo.Metrics = make([]AliyunMetric, 0)
	for key, countObj := range succCountMap {
		dimessionObj := parseSuccessCountDimension(key)
		if nil == dimessionObj {
			logrus.WithFields(logrus.Fields{"key":key}).Warn("report success count dimession key abnormal")
			continue
		}
		if "" != dimessionObj.SName {
			dimessionObj.SName += "_" + monitorObj.conf.EnvironmenType
		}
		if "" != dimessionObj.TName {
			dimessionObj.TName += "_" + monitorObj.conf.EnvironmenType
		}

		var metric AliyunMetric
		metric.MetricName = dimessionObj.getMetricName()
		metric.Value = countObj.counter
		metric.Timestamp = reportTime.UnixNano() / 1e6
		metric.Dimensions = dimessionObj
		reportInfo.Metrics = append(reportInfo.Metrics, metric)
	}
	if len(reportInfo.Metrics) > 0 {
		sendRequestToAliMonitor(&reportInfo)
	}
}

func reportFailedCountToAliyun(failedCountMap map[string]countInfo, reportTime time.Time) {
	var reportInfo AliyunMonitorInfo
	reportInfo.Namespace = namespace
	reportInfo.UserId = uid
	reportInfo.Metrics = make([]AliyunMetric, 0)
	for key, countObj := range failedCountMap {
		dimessionObj := parseFailedCountDimension(key)
		if nil == dimessionObj {
			logrus.WithFields(logrus.Fields{"key":key}).Warn("report failed count dimession key abnormal")
			continue
		}
		if "" != dimessionObj.SName {
			dimessionObj.SName += "_" + monitorObj.conf.EnvironmenType
		}
		if "" != dimessionObj.TName {
			dimessionObj.TName += "_" + monitorObj.conf.EnvironmenType
		}

		var metric AliyunMetric
		metric.MetricName = dimessionObj.getMetricName()
		metric.Value = countObj.counter
		metric.Timestamp = reportTime.UnixNano() / 1e6
		metric.Dimensions = dimessionObj
		reportInfo.Metrics = append(reportInfo.Metrics, metric)
	}
	if len(reportInfo.Metrics) > 0 {
		sendRequestToAliMonitor(&reportInfo)
	}
}


func reportSuccAvgTimeToAliyun(succAvgTimeMap map[string]countInfo, reportTime time.Time) {
	var reportInfo AliyunMonitorInfo
	reportInfo.Namespace = namespace
	reportInfo.UserId = uid
	reportInfo.Metrics = make([]AliyunMetric, 0)
	for key, countObj := range succAvgTimeMap {
		dimessionObj := parseSuccessAvgTimeDimension(key)
		if nil == dimessionObj {
			logrus.WithFields(logrus.Fields{"key":key}).Warn("report success avg time dimession key abnormal")
			continue
		}
		if "" != dimessionObj.SName {
			dimessionObj.SName += "_" + monitorObj.conf.EnvironmenType
		}
		if "" != dimessionObj.TName {
			dimessionObj.TName += "_" + monitorObj.conf.EnvironmenType
		}

		var metric AliyunMetric
		metric.MetricName = dimessionObj.getMetricName()
		metric.Value = countObj.sum / countObj.counter			//平均值
		metric.Timestamp = reportTime.UnixNano() / 1e6
		metric.Dimensions = dimessionObj
		reportInfo.Metrics = append(reportInfo.Metrics, metric)
	}
	if len(reportInfo.Metrics) > 0 {
		sendRequestToAliMonitor(&reportInfo)
	}
}

func reportFailedAvgTimeToAliyun(failedAvgTimeMap map[string]countInfo, reportTime time.Time) {
	var reportInfo AliyunMonitorInfo
	reportInfo.Namespace = namespace
	reportInfo.UserId = uid
	reportInfo.Metrics = make([]AliyunMetric, 0)
	for key, countObj := range failedAvgTimeMap {
		dimessionObj := parseFailedAvgTimeDimension(key)
		if nil == dimessionObj {
			logrus.WithFields(logrus.Fields{"key":key}).Warn("report failed avg time dimession key abnormal")
			continue
		}
		if "" != dimessionObj.SName {
			dimessionObj.SName += "_" + monitorObj.conf.EnvironmenType
		}
		if "" != dimessionObj.TName {
			dimessionObj.TName += "_" + monitorObj.conf.EnvironmenType
		}

		var metric AliyunMetric
		metric.MetricName = dimessionObj.getMetricName()
		metric.Value = countObj.sum / countObj.counter			//平均值
		metric.Timestamp = reportTime.UnixNano() / 1e6
		metric.Dimensions = dimessionObj
		reportInfo.Metrics = append(reportInfo.Metrics, metric)
	}
	if len(reportInfo.Metrics) > 0 {
		sendRequestToAliMonitor(&reportInfo)
	}
}