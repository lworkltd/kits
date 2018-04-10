package report

import (
	"sync"
	"time"
)

// Reporter 上报服务器
type Reporter interface {
	Report(*Metric)
}

// Metric 上报的数据
type Metric struct {
	service string
	reqName string
	delay   time.Duration
	code    string
}

var reportMetricPool *sync.Pool

func init() {
	reportMetricPool = &sync.Pool{
		New: func() interface{} {
			return &Metric{}
		},
	}
}

// Service 服务名称
func (rm *Metric) Service(sn string) *Metric {
	rm.service = sn
	return rm
}

// Delay 请求延迟
func (rm *Metric) Delay(delay time.Duration) *Metric {
	rm.delay = delay
	return rm
}

// ReqName 请求名称
func (rm *Metric) ReqName(reqName string) *Metric {
	rm.reqName = reqName
	return rm
}

// Failed 请求失败
func (rm *Metric) Failed(code string) *Metric {
	rm.code = code
	return rm
}

// Succ 请求成功
func (rm *Metric) Succ() *Metric {
	rm.code = "succ"
	return rm
}

// Clone 服务名称
func (rm *Metric) Clone() *Metric {
	newRm := NewMetric()
	copyReportMetric(rm, newRm)
	return newRm
}

func (rm *Metric) reset() {
	rm.service = ""
	rm.code = ""
	rm.delay = 0
	rm.reqName = ""
}

// NewMetric 新建一个数据
func NewMetric() *Metric {
	rm := reportMetricPool.Get().(*Metric)
	rm.reset()
	return rm
}

func copyReportMetric(from, to *Metric) {
	*to = *from
}
