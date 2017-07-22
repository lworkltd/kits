package schedule

import (
	"time"
)

type Period int8

const (
	Undefined Period = 0 // 未定义
	Hour      Period = 1 // 小时
	Day       Period = 2 // 天
	Week      Period = 3 // 周
	Mouth     Period = 4 // 月
)

// 调度器的时间相关执行基准时区
// 比如以+8时区为基准下，每天中午触发执行条件时，实际是在UTC时间的20点
var location = time.UTC

// Scheduler 是一个调度器，他规定了什么时刻做某事
// Scheduler 有的策略是互斥的，比如你不能即规定每天固定时间调度，并规定每隔5分钟执行
type Scheduler interface {
	// 定义了循环的调度的时间策略
	Every(time.Duration) Scheduler

	// 定义了定时执行的时间策略
	// TODO:还没有做，请期待着或者自己完善
	AtTime(Period, time.Duration) Scheduler

	// 定义条件执行的调度
	If(func() bool) Scheduler

	// 定义了执行次数的调度策略
	Count(int) Scheduler

	// 定义了延迟执行的调度策略
	Delay(time.Duration) Scheduler

	// 启动调度策略
	Start(func()) Scheduler

	// 安全调度
	// 如果发生panic，将不会退出调度
	Safety() Scheduler

	// 强制退出调度循环
	Close() Scheduler

	// 条件退出调度循环
	// 当条件触发后，将退出调度
	CloseIf(func() bool) Scheduler
}

// New 创建一个调度器
func New() Scheduler {
	return &schedulerImpl{}
}
