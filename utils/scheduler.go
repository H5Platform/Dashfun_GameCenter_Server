package utils

import (
	"context"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Task 表示一个可定时执行的任务
type Task struct {
	Name     string
	Interval time.Duration
	Callback func()
	cancel   context.CancelFunc
}

// Scheduler 管理多个带 ctx 的定时任务
type Scheduler struct {
	ctx   context.Context
	mu    sync.Mutex
	tasks map[string]*Task
}

// NewScheduler 创建一个新的调度器
func NewScheduler(ctx context.Context) *Scheduler {
	return &Scheduler{
		ctx:   ctx,
		tasks: make(map[string]*Task),
	}
}

// Add 添加一个新任务（如果已存在相同名称的任务会返回 false）
func (s *Scheduler) Add(name string, interval time.Duration, callback func()) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[name]; exists {
		return false
	}

	taskCtx, cancel := context.WithCancel(s.ctx)
	task := &Task{
		Name:     name,
		Interval: interval,
		Callback: callback,
		cancel:   cancel,
	}

	s.tasks[name] = task
	go s.runTask(taskCtx, task)
	zap.S().Infow("Schedule Task Added", "name", name, "interval", interval)
	return true
}

// Remove 删除指定名称的任务
func (s *Scheduler) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[name]; exists {
		task.cancel()
		delete(s.tasks, name)
	}
}

// Replace 替换已有任务（如果存在则先删除，再添加）
func (s *Scheduler) Replace(name string, interval time.Duration, callback func()) {
	s.Remove(name)
	s.Add(name, interval, callback)
}

// runTask 启动具体的任务逻辑
func (s *Scheduler) runTask(ctx context.Context, task *Task) {
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			task.Callback()
		case <-ctx.Done():
			zap.S().Infow("Schedule Task Stopped", "name", task.Name)
			return
		}
	}
}

// StopAll 所有任务
func (s *Scheduler) StopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, task := range s.tasks {
		task.cancel()
	}
	s.tasks = make(map[string]*Task)
}
