package task

import (
	"time"

	"github.com/google/uuid"
)

// ITask 任务标准
type ITask interface {
	// TaskNo 任务编号
	TaskNo() string
	// Delay 延迟时间
	Delay() time.Duration
	// IsCycle 是否周期
	IsCycle() bool
	// Interval 执行间隔
	Interval() int
	// Execute 任务执行
	Execute() func() error
}

// 任务
type task struct {
	o *options
}

// New 创建
func New(opts ...Option) (ITask, error) {
	o := &options{
		no:         uuid.NewString(),
		createTime: time.Now(),
	}

	for _, opt := range opts {
		if err := opt(o); nil != err {
			return nil, err
		}
	}

	return &task{
		o: o,
	}, nil
}

// Copy 复制任务 叠加可选属性
func Copy(it ITask, opts ...Option) (ITask, error) {
	t, ok := it.(*task)
	if !ok {
		return New(opts...)
	}

	for _, opt := range opts {
		if err := opt(t.o); nil != err {
			return nil, err
		}
	}

	return &task{
		o: t.o,
	}, nil
}

// TaskNo 任务编号
func (t *task) TaskNo() string {
	return t.o.no
}

// Delay 延迟时间
func (t *task) Delay() time.Duration {
	return t.o.delayTime.Sub(t.o.createTime)
}

// Execute 执行
func (t *task) Execute() func() error {
	return t.o.callback
}

// IsCycle 是否周期
func (t *task) IsCycle() bool {
	return t.o.cycle
}

// Interval 执行间隔
func (t *task) Interval() int {
	return t.o.interval
}
