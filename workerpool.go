package workerpool

import (
	"context"
	"errors"
	"sync"
)

var (
	WorkerPoolStatusStopped  = 0
	WorkerPoolStatusRunning  = 1
	WorkerPoolStatusStopping = 2

	WorkerPoolBadCommandError = errors.New("Operation aborted! (could cause invalid state)")
)

// WorkerPool main engine rensponsible for managing workers
type WorkerPool struct {
	sync.Mutex
	status   int
	tasks    chan Task
	capacity int
	quit     context.CancelFunc
	wg       sync.WaitGroup
}

// Run starts the workers
func (wp *WorkerPool) Run(ctx context.Context) error {
	wp.Lock()
	defer wp.Unlock()

	if wp.status != WorkerPoolStatusStopped {
		return WorkerPoolBadCommandError
	}

	ctx, cancel := context.WithCancel(ctx)
	wp.quit = cancel

	for i := 0; i < wp.capacity; i++ {
		workerCtx, _ := context.WithCancel(ctx)
		go func(ctx context.Context, work chan Task, wg *sync.WaitGroup, index int) {
			for task := range work {
				task.Run(ctx)
				wg.Done()
			}
		}(workerCtx, wp.tasks, &wp.wg, i)
	}

	wp.status = WorkerPoolStatusRunning
	return nil
}

// Execute appends work to the workers queue
func (wp *WorkerPool) Execute(task Task) error {
	wp.Lock()
	defer wp.Unlock()

	if wp.status != WorkerPoolStatusRunning {
		return WorkerPoolBadCommandError
	}

	wp.wg.Add(1)
	wp.tasks <- task
	return nil
}

// Shutdown sends cancel signal to workers in the queue
func (wp *WorkerPool) Shutdown() error {
	wp.Lock()
	defer wp.Unlock()

	if wp.status != WorkerPoolStatusRunning {
		return WorkerPoolBadCommandError
	}
	wp.status = WorkerPoolStatusStopping
	wp.quit()
	close(wp.tasks)
	wp.wg.Wait()
	wp.status = WorkerPoolStatusStopped
	return nil
}

// New worker pool constructor
func New(capacity int) *WorkerPool {
	return &WorkerPool{
		wg:       sync.WaitGroup{},
		capacity: capacity,
		tasks:    make(chan Task),
		status:   WorkerPoolStatusStopped,
	}
}

// Task interface of worker tasks
type Task interface {
	Run(ctx context.Context) error
}
