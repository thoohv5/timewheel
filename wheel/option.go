package wheel

import (
	"time"
)

// Option 可选参数方法
type Option func(opt *options) error

// 轮模型参数
type options struct {
	wheel IWheel

	// 时间间隔
	tick time.Duration
	// 间隔数量
	bucketsNum int
}

// WithTick 时间间隔
func WithTick(tick time.Duration) Option {
	return func(o *options) error {
		if tick < time.Microsecond {
			return ErrIllegalTick
		}
		o.tick = tick
		return nil
	}
}

// WithBucketsNum 间隔数量
func WithBucketsNum(bucketsNum int) Option {
	return func(o *options) error {
		if bucketsNum <= 0 {
			return ErrIllegalBucketsNum
		}
		o.bucketsNum = bucketsNum
		return nil
	}
}

// WithWheel 上级时间轮
func WithWheel(wheel IWheel) Option {
	return func(o *options) error {
		o.wheel = wheel
		return nil
	}
}
