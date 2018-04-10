package report

import (
	"testing"
	"time"
)

func TestStatsdGrpcReporterReport(t *testing.T) {
	type args struct {
		rm *Metric
	}
	tests := []struct {
		name string
		r    Reporter
		args args
	}{
		{
			name: "failed",
			r: func() Reporter {
				reporter, _ := NewStatsdGrpcReporter("127.0.0.1:9091")
				return reporter
			}(),
			args: args{
				rm: NewMetric().
					Delay(time.Second).
					Failed("mt4_101").
					ReqName("AddRequest").
					Service("Calculator"),
			},
		},
		{
			name: "succ",
			r: func() Reporter {
				reporter, _ := NewStatsdGrpcReporter("127.0.0.1:9091")
				return reporter
			}(),
			args: args{
				rm: NewMetric().
					Delay(time.Second).
					Succ().
					ReqName("AddRequest").
					Service("Calculator"),
			},
		},

		{
			name: "clone",
			r: func() Reporter {
				reporter, _ := NewStatsdGrpcReporter("127.0.0.1:9091")
				return reporter
			}(),
			args: args{
				rm: NewMetric().Clone().
					Delay(time.Second).
					Succ().
					ReqName("AddRequest").
					Service("Calculator"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Report(tt.args.rm)
		})
	}
}
