package slot

import (
	"container/list"
	"sync"

	"github.com/thoohv5/timewheel/task"
)

// ISlot 槽模型标准
type ISlot interface {
	// Add 添加任务
	Add(task task.ITask) (taskNo string, err error)
	// Delete 删除任务
	Delete(no string) (err error)
	// Traverse 遍历任务
	Traverse(tf func(t task.ITask) error) (err error)
}

type slot struct {
	// 关联task, key: taskno, value: task e
	m sync.Map
	// 关联task
	e *list.List
}

// New 创建
func New() ISlot {
	return &slot{
		e: list.New(),
	}
}

func (s *slot) Add(t task.ITask) (no string, err error) {
	no = t.TaskNo()
	s.m.Store(no, t)
	s.e.PushBack(t)
	return
}

func (s *slot) Delete(no string) (err error) {
	t, ok := s.m.Load(no)
	if !ok {
		return
	}
	s.e.Remove(&list.Element{
		Value: t,
	})
	s.m.Delete(no)
	return
}

func (s *slot) Traverse(tf func(t task.ITask) error) (err error) {
	for node := s.e.Front(); node != nil; node = node.Next() {
		if err = tf(node.Value.(task.ITask)); nil != err {
			return
		}
	}
	return
}
