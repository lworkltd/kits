package schedule

import (
	"runtime/debug"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
)

type TimeStrategy int8

const (
	TimeStrategyEvery  TimeStrategy = 0
	TimeStrategyAtTime TimeStrategy = 1
)

type nextWait interface {
	wait(bool)
}

// interval是一个定期等待器
type atTime struct {
	period Period
	offset time.Duration
	local  *time.Location
}

func (me *atTime) wait(_ bool) {
	//TODO:实现定时触发
	panic("not implement")
}

// interval是一个定时等待器
type interval struct {
	dur time.Duration
}

func (me *interval) wait(_ bool) {
	<-time.After(me.dur)
}

// schedulerImpl 是Scheduler的实现
type schedulerImpl struct {
	uuid       string
	startDelay time.Duration
	nextWait   nextWait
	count      int
	cond       func() bool
	closeIf    func() bool
	safety     bool
	looping    bool
	f          func()
}

// 定义了循环的调度的时间策略
func (me *schedulerImpl) Every(dur time.Duration) Scheduler {
	if me.nextWait != nil {
		panic("time strategy already set")
	}

	me.nextWait = &interval{
		dur: dur,
	}

	return me
}

// 定义了定时执行的时间策略
func (me *schedulerImpl) AtTime(p Period, at time.Duration) Scheduler {
	me.nextWait = &atTime{
		period: p,
		offset: at,
		local:  location,
	}
	return me
}

// 定义条件执行的调度
func (me *schedulerImpl) If(cond func() bool) Scheduler {
	me.cond = cond
	return me
}

// 定义了执行次数的调度策略
func (me *schedulerImpl) Count(c int) Scheduler {
	me.count = c
	return me
}

// 定义了延迟执行的调度策略
func (me *schedulerImpl) Delay(delay time.Duration) Scheduler {
	me.startDelay = delay
	return me
}

// 启动调度策略
func (me *schedulerImpl) Start(f func()) Scheduler {
	me.uuid = uuid.New().String()
	me.f = f
	me.looping = true
	go me.loop()

	return me
}

// 安全调度
// 如果发生panic，将不会退出调度
func (me *schedulerImpl) Safety() Scheduler {
	me.safety = true
	return me
}

// 强制退出调度循环
func (me *schedulerImpl) Close() Scheduler {
	me.looping = false
	return me
}

// 条件退出调度循环
// 当条件触发后，将退出调度
func (me *schedulerImpl) CloseIf(f func() bool) Scheduler {
	me.closeIf = f
	return me
}

func (me *schedulerImpl) keepRunning() bool {
	if !me.looping {
		return false
	}

	if me.closeIf != nil && me.closeIf() {
		return false
	}

	if me.count <= 0 {
		return false
	}

	return true
}

func (me *schedulerImpl) loop() {
	if me.safety {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithFields(logrus.Fields{
					"uuid":  me.uuid,
					"stack": debug.Stack(),
				}).Error("Scheduler panic")
				go me.loop()
			}
		}()
	}

	if me.startDelay > 0 {
		<-time.After(me.startDelay)
		me.startDelay = 0
	}

	for me.keepRunning() {
		me.nextWait.wait(me.keepRunning())
		if me.cond != nil && !me.cond() {
			continue
		}

		me.f()
		me.count--
	}
}
