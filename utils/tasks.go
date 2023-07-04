package utils

import (
	"sync"
	"time"
)

var tasks *TaskTimer

// TaskTimer linter
type TaskTimer struct {
	tasks map[string]*time.Timer
	mutex sync.RWMutex
}

func init() {
	tasks = NewTaskTimer()
}

// NewTaskTimer linter
func NewTaskTimer() *TaskTimer {
	return &TaskTimer{
		tasks: make(map[string]*time.Timer),
	}
}

func (t *TaskTimer) remove(id string) {
	if tks := t.get(id); tks != nil {
		tks.Stop()
		t.delete(id)
	}
}

func (t *TaskTimer) delete(id string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.tasks, id)
}

func (t *TaskTimer) add(id string, tks *time.Timer) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tasks[id] = tks
}

func (t *TaskTimer) get(id string) *time.Timer {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.tasks[id]
}

// AddTask linter
func AddTask(id string, tks *time.Timer) {
	tasks.remove(id)
	tasks.add(id, tks)
}

// RemoveTask linter
func RemoveTask(id string) {
	tasks.remove(id)
}
