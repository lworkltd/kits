package log

import (
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestJSONFormatter_Format(t *testing.T) {
	type args struct {
		entry *logrus.Entry
	}
	tests := []struct {
		name    string
		f       *JSONFormatter
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.Format(tt.args.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONFormatter.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JSONFormatter.Format() = %v, want %v", got, tt.want)
			}
		})
	}
}
