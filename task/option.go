package task

import "time"

// Option 可选参数方法
type Option func(o *options) error

// 任务参数
type options struct {
	// 编号
	no string
	// 间隔次数
	interval int

	// 创建时间
	createTime time.Time
	// 延迟时间
	delayTime time.Time
	// 数据
	callback func() error

	// 重试次数
	retry int
	// 重试时间
	retryTime time.Time

	// 周期任务
	cycle bool
	// 执行次数
	times int
}

// WithNo 编号
func WithNo(no string) Option {
	return func(o *options) error {
		o.no = no
		return nil
	}
}

// WithInterval 间隔次数
func WithInterval(interval int) Option {
	return func(o *options) error {
		o.interval = interval
		return nil
	}
}

// WithDelayTime 延迟时间
func WithDelayTime(delayTime time.Time) Option {
	return func(o *options) error {
		o.delayTime = delayTime
		return nil
	}
}

// WithCallBack 回调方法
func WithCallBack(callback func() error) Option {
	return func(o *options) error {
		o.callback = callback
		return nil
	}
}

// WithRetry 重试次数
func WithRetry() Option {
	return func(o *options) error {
		o.retry += 1
		return nil
	}
}

// WithRetryTime 重试时间
func WithRetryTime(retryTime time.Time) Option {
	return func(o *options) error {
		o.retryTime = retryTime
		return nil
	}
}

// WithCycle 周期任务
func WithCycle() Option {
	return func(o *options) error {
		o.cycle = true
		return nil
	}
}

// WithTimes 执行次数
func WithTimes() Option {
	return func(o *options) error {
		o.times++
		return nil
	}
}
