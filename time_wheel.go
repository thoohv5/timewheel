package timewheel

import (
	"sync"
	"time"

	"github.com/thoohv5/timewheel/task"
	"github.com/thoohv5/timewheel/wheel"
)

// ITimeWheel 时间轮标准
type ITimeWheel interface {
	// Start 开始
	Start()
	// Stop 停止
	Stop()
	// AddTask 添加任务
	AddTask(ts time.Duration, callback func() error, opts ...task.Option) (string, error)
	// AddCycleTask 添加周期任务
	AddCycleTask(ts time.Duration, callback func() error, opts ...task.Option) (string, error)
}

// 时间轮
type timeWheel struct {
	tick      time.Duration
	ticker    *time.Ticker
	tickQueue chan time.Time

	stopC chan struct{}

	onceStart sync.Once
	exited    bool

	wheel wheel.IWheel
}

const (
	defaultTickTime     = time.Second
	defaultTaskQueueNum = 10
)

// New 创建
func New(opts ...wheel.Option) (ITimeWheel, error) {
	w, err := wheel.New(opts...)
	if nil != err {
		return nil, err
	}
	return &timeWheel{
		tick:      defaultTickTime,
		tickQueue: make(chan time.Time, defaultTaskQueueNum),
		stopC:     make(chan struct{}),
		wheel:     w,
	}, nil
}

// Start 开始
func (tw *timeWheel) Start() {
	tw.onceStart.Do(
		func() {
			tw.ticker = time.NewTicker(tw.tick)
			go tw.scheduler()
			go tw.trigger()
		},
	)
}

// Stop 停止
func (tw *timeWheel) Stop() {
	tw.stopC <- struct{}{}
}

// AddTask 添加任务
func (tw *timeWheel) AddTask(ts time.Duration, callback func() error, opts ...task.Option) (string, error) {
	return tw.wheel.CreateTask(ts, append([]task.Option{task.WithCallBack(callback)}, opts...)...)
}

// AddCycleTask 添加周期任务
func (tw *timeWheel) AddCycleTask(ts time.Duration, callback func() error, opts ...task.Option) (string, error) {
	return tw.wheel.CreateTask(ts, append([]task.Option{task.WithCycle(), task.WithCallBack(callback)}, opts...)...)
}

// trigger 触发
func (tw *timeWheel) trigger() {
	if tw.tickQueue != nil {
		return
	}
	for !tw.exited {
		select {
		case <-tw.ticker.C:
			select {
			case tw.tickQueue <- time.Now():
			default:
				panic("raise long time blocking")
			}
		}
	}
}

// scheduler 计划程序
func (tw *timeWheel) scheduler() {
	queue := tw.ticker.C
	if tw.tickQueue == nil {
		queue = tw.tickQueue
	}

	for {
		select {
		case <-queue:
			tw.wheel.Roll()

		case <-tw.stopC:
			tw.exited = true
			tw.ticker.Stop()
			return
		}
	}
}
