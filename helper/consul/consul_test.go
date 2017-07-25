package consul

import (
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name string
		want *Client
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
