package code

import (
	"fmt"
	"testing"
)

func TestErrorString(t *testing.T) {
	tests := []struct {
		name string
		err  Error
		want string
	}{
		{
			err: &errorImpl{
				message: "error message",
				mcode:   "MCODE",
				code:    10,
			},
			want: "MCODE,error message",
		},
		{
			err: &errorImpl{
				message: "error message",
				mcode:   "",
				code:    10,
			},
			want: "10,error message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmt.Sprintf("%v", tt.err); got != tt.want {
				t.Errorf("errorImpl.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
