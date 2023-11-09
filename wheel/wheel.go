package wheel

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/thoohv5/timewheel/slot"
	"github.com/thoohv5/timewheel/task"
	"github.com/thoohv5/timewheel/util/sync/errgroup"
)

// IWheel 轮模型标准
type IWheel interface {
	// CreateTask 创建任务
	CreateTask(ts time.Duration, opts ...task.Option) (string, error)
	// Roll 推动轮转动
	Roll() error
}

// 轮模型实体
type wheel struct {
	// 参数
	o *options

	// 当前指针
	index int
	// 关联slot
	m sync.Map

	// 上级轮，不存在时为nil
	superior IWheel
}

const (
	// 默认间隔时间
	defaultTickTime = time.Second
	// 默认间隔数量
	defaultBucketsNum = 60
)

var (
	ErrIllegalTick          = errors.New("illegal tick")
	ErrIllegalBucketsNum    = errors.New("illegal buckets num")
	ErrIllegalDelayTime     = errors.New("illegal delay time")
	ErrIllegalSuperiorWheel = errors.New("illegal superior wheel")
)

func New(opts ...Option) (IWheel, error) {
	o := &options{
		tick:       defaultTickTime,
		bucketsNum: defaultBucketsNum,
	}

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	return &wheel{
		o:        o,
		superior: o.wheel,
	}, nil
}

// CreateTask 创建任务
func (w *wheel) CreateTask(ts time.Duration, opts ...task.Option) (taskNo string, err error) {
	interval := int(ts / w.o.tick)
	// 查找所属槽
	s, err := w.findSlot(interval)
	if nil != err {
		return
	}

	// 创建任务
	t, err := task.New(append(
		[]task.Option{
			task.WithInterval(interval),
		}, opts...)...)
	if nil != err {
		return
	}

	// 向槽中添加任务
	taskNo, err = s.Add(t)
	if nil != err {
		return
	}

	return
}

// Roll 推动轮转动
func (w *wheel) Roll() (err error) {
	defer func() {
		if err != nil {
			return
		}
		// 自动增加
		w.index++
		// 向上传递
		err = w.delivery()
		if err != nil {
			return
		}
		return
	}()

	// 获取当前的槽
	ms, ok := w.m.Load(w.index)
	if !ok {
		return
	}
	s := ms.(slot.ISlot)

	// 处理槽中的任务
	eg := errgroup.WithContext(context.Background())
	err = s.Traverse(func(t task.ITask) (err error) {
		// 删除任务
		if err = s.Delete(t.TaskNo()); err != nil {
			return
		}
		eg.Go(func(ctx context.Context) error {
			return w.execute(t)
		})
		return
	})
	if nil != err {
		return
	}
	if err = eg.Wait(); nil != err {
		return
	}

	return
}

// copy 复制任务
func (w *wheel) copy(t task.ITask, interval int, opts ...task.Option) (taskNo string, err error) {
	// 查找所属槽
	s, err := w.findSlot(interval)
	if nil != err {
		return
	}

	// 追加任务属性
	t, err = task.Copy(t, opts...)
	if nil != err {
		return
	}

	// 向槽中添加任务
	taskNo, err = s.Add(t)
	if nil != err {
		return
	}

	return
}

// 寻找指定的槽
func (w *wheel) findSlot(interval int) (s slot.ISlot, err error) {
	curWheel := w
	taskTime := time.Duration(interval) * curWheel.o.tick
	wholeTime := curWheel.o.tick * time.Duration(curWheel.o.bucketsNum)
	for taskTime > wholeTime {
		// 延迟时间超出所有轮的长度
		if curWheel.superior == nil {
			err = ErrIllegalDelayTime
			return
		}

		// 上级轮不是指定类型的轮
		cw, ok := curWheel.superior.(*wheel)
		if !ok {
			err = ErrIllegalSuperiorWheel
			return
		}

		curWheel = cw
	}

	// 计算延迟时间对应的位置
	key := (interval + curWheel.index) % curWheel.o.bucketsNum

	// 通过位置查找当前的槽
	ms, ok := curWheel.m.Load(key)
	if !ok {
		// 新添加的槽，需要记录到对应的轮中
		s = slot.New()
		curWheel.m.Store(key, s)
	} else {
		s = ms.(slot.ISlot)
	}

	return
}

// 推动轮向上传递，同时把满足添加的任务，往下级轮转移
func (w *wheel) delivery() error {
	curWheel := w

	for curWheel.index >= w.o.bucketsNum {
		curWheel.index %= w.o.bucketsNum
		// 延迟时间超出所有轮的长度
		if curWheel.superior != nil {
			return ErrIllegalDelayTime
		}

		// 上级轮不是指定类型的轮
		cw, ok := curWheel.superior.(*wheel)
		if !ok {
			return ErrIllegalSuperiorWheel
		}
		preWheel := curWheel
		curWheel = cw
		// 通过位置查找当前的槽
		ms, ok := curWheel.m.Load(curWheel.index)
		if ok {
			s := ms.(slot.ISlot)
			if err := s.Traverse(func(t task.ITask) error {
				if err := s.Delete(t.TaskNo()); err != nil {
					return err
				}
				key := t.Delay() - time.Duration(curWheel.o.bucketsNum*curWheel.index)
				preWheel.m.Store(key, t)
				return nil
			}); err != nil {
				return err
			}
		}
		curWheel.index++
	}

	return nil
}

// 指定任务，如果任务失败，重新添加任务，默认延迟10s
func (w *wheel) execute(t task.ITask) (err error) {
	if t.IsCycle() {
		// 添加任务
		_, err = w.copy(t, t.Interval(), task.WithTimes())
		if err != nil {
			return err
		}
	}

	// 执行任务
	if err = t.Execute()(); err == nil {
		return
	}

	// 添加任务
	_, err = w.copy(t, int((10*time.Second)/w.o.tick), task.WithRetry(), task.WithRetryTime(time.Now()))
	if err != nil {
		return err
	}

	return
}
