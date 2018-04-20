package report

import (
	"testing"
	"time"

	"github.com/lworkltd/kits/service/monitor"
)

func TestMonitorReporter_Report(t *testing.T) {
	type args struct {
		reqInterface string
		reqService   string
		fromHost     string
		result       string
		delay        time.Duration
	}
	monitor.Init(&monitor.MonitorConf{
		EnableReport: true,
		ReportAddr:   "domain.notexist.com:8080",
	})
	tests := []struct {
		name     string
		reporter *MonitorReporter
		args     args
	}{
		{
			name:     "succ",
			reporter: &MonitorReporter{},
			args: args{
				reqInterface: "AddRequest",
				reqService:   "Calculator",
				fromHost:     "127.0.0.1",
				result:       "",
				delay:        time.Second,
			},
		},
		{
			name:     "error",
			reporter: &MonitorReporter{},
			args: args{
				reqInterface: "AddRequest",
				reqService:   "Calculator",
				fromHost:     "127.0.0.1",
				result:       "EORROR",
				delay:        time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.reporter.Report(tt.args.reqInterface, tt.args.reqService, tt.args.fromHost, tt.args.result, tt.args.delay)
		})
	}
}
