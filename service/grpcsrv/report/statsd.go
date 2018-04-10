package report

import (
	"fmt"
	"net"
	"strings"
)

// StatsdGrpcReporter 上报给Statsd
type StatsdGrpcReporter struct {
	conn net.Conn
}

// Report 上报
func (r *StatsdGrpcReporter) Report(rm *Metric) {
	data := []string{
		"grpcsrv",
		rm.service,
		rm.reqName,
		rm.code,
		fmt.Sprintf("%v", rm.delay),
	}

	fmt.Fprintf(r.conn, strings.Join(data, "."))
}

// NewStatsdGrpcReporter 创建一个发送Statsd数据的上报器
func NewStatsdGrpcReporter(udpAddr string) (*StatsdGrpcReporter, error) {
	conn, err := net.Dial("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return &StatsdGrpcReporter{
		conn: conn,
	}, nil
}
