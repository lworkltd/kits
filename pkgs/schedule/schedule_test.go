package schedule

import (
	"sync"
	"testing"
	"time"
)
func Test(t *testing.T) {
	
}
func TestNew(t *testing.T) {
	c := 0
	cond := func() bool {
		c++
		//fmt.Printf("-------- %d\n", c)
		return c < 3 || c > 5
	}
	closeCond := func() bool {
		return c == 10
	}
	tests := []struct {
		name string
		want Scheduler
	}{
		{want: &schedulerImpl{
			cond:       cond,
			count:      10,
			closeIf:    closeCond,
			startDelay: time.Second,
			safety:     true,
		}},
	}
	wg := sync.WaitGroup{}
	wg.Add(7)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New().
				If(cond).
				Count(10).
				Safety().
				CloseIf(closeCond).
				Delay(time.Millisecond * 20)
			got.Every(time.Millisecond * 50).Start(func() {
				//fmt.Printf("########### %d\n", c)
				wg.Done()
			})
		})
	}
	wg.Wait()
	if c != 10 {
		t.Errorf("expect 10 got %v\n", c)
	}
}
