package parallel

import (
	"runtime/debug"

	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
)

// Parallel 是将串行的任务并行化的调度器
type Parallel interface {
	// 启动并行化
	Start() Parallel
	// 停止并行化
	Stop()
	// 投放任务
	Put(Task)
}

// Task is Parallel的最小处理单元
type Task interface {
	// 定义任务是否可以抛弃，当并行化效率低于低于投递效率时将根据这个字段判断
	Abandonable() bool
	// 执行函数
	Deal()
}

// ParallelImpl 是并行化接口的实现
type ParallelImpl struct {
	uuid        string
	tasks       chan Task
	maxRoutines int
	looping     bool
}

func (parallel *ParallelImpl) loop() {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"uuid":  parallel.uuid,
				"stack": debug.Stack(),
			}).Error("Parallel panic")

			go parallel.loop()
		}
	}()

	for parallel.looping {
		select {
		case task := <-parallel.tasks:
			if len(parallel.tasks) == parallel.maxRoutines {
				if task.Abandonable() {
					continue
				}
			}

			task.Deal()
		}
	}
}

// Put 投放任务
func (me *ParallelImpl) Put(task Task) {
	me.tasks <- task
}

// Start 启动并行化
func (me *ParallelImpl) Start() Parallel {
	me.looping = true
	for index := 0; index < me.maxRoutines; index++ {
		go me.loop()
	}

	return me
}

// Stop 启动并行化
func (me *ParallelImpl) Stop() {
	me.looping = false
	close(me.tasks)
}

// New 创建了一个并行化对象，max是最大并行化数量
func New(max int) Parallel {
	return &ParallelImpl{
		uuid:        uuid.New().String(),
		maxRoutines: max,
		tasks:       make(chan Task, 1000),
	}
}
