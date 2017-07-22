package parallel

import (
	"sync"
	"testing"

	"lwork.com/kits/utils/co"
)

var counter co.Int64
var wg sync.WaitGroup

type TestTask struct {
	Number int
}

func (task *TestTask) Abandonable() bool {
	return false
}

func (task *TestTask) Deal() {
	counter.Add(1)
	wg.Done()
}

func TestParallel(t *testing.T) {
	counter = 0
	p := New(100)
	p.Start()
	wg.Add(100)
	for i := 0; i < 100; i++ {
		p.Put(&TestTask{})
	}
	wg.Wait()

	if counter.Get() != 100 {
		t.Errorf("counter value is %d", counter)
	}
}
