package wrap

import (
	"reflect"
	"testing"

	"github.com/lworkltd/kits/service/restful/code"
)

func TestSnowSlide_Check(t *testing.T) {
	tests := []struct {
		name      string
		snowslide *SnowSlide
		want      code.Error
	}{
		{
			snowslide: &SnowSlide{
				LimitCnt: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for index := 0; index < 9; index++ {
				if got := tt.snowslide.Check(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("SnowSlide.Check() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}
