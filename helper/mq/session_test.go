package mq

import (
	"reflect"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TesMakePublishing(t *testing.T) {
	now := time.Now()
	type args struct {
		options []PublishOption
	}
	tests := []struct {
		name string
		args args
		want *amqp.Publishing
	}{
		{
			name: "appid",
			args: args{
				[]PublishOption{
					OptionAppId("abc123"),
					OptionTimestamp(now),
				},
			},
			want: &amqp.Publishing{
				AppId:     "abc123",
				Timestamp: now,
			},
		},
		{
			name: "content-type",
			args: args{
				[]PublishOption{
					OptionContentType("type1"),
					OptionTimestamp(now),
				},
			},
			want: &amqp.Publishing{
				ContentType: "type1",
				Timestamp:   now,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makePublishing(nil, tt.args.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makePublishing() = %v, want %v", got, tt.want)
			}
		})
	}
}
