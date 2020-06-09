package workerpool

import (
	"context"
	"errors"
	"sync"
)

var (
    WorkerPoolStatusStopped = 0
    WorkerPoolStatusRunning = 1
    WorkerPoolStatusStopping = 2

    WorkerPoolOperationAbortedError = errors.New("Operation aborted! (could cause invalid state)")
)

type WorkerPool struct {
	sync.Mutex
	status   int
	tasks    chan Task
	capacity int
	q        context.CancelFunc
	wg       sync.WaitGroup
}

func (wp *WorkerPool) AddWork(task Task) error {
	wp.Lock()
	defer wp.Unlock()

	if wp.status != WorkerPoolStatusRunning {
		return WorkerPoolOperationAbortedError
	}

	wp.wg.Add(1)
	wp.tasks <- task
	return nil
}

func (wp *WorkerPool) Start(ctx context.Context) error {
	wp.Lock()
	defer wp.Unlock()

	if wp.status != WorkerPoolStatusStopped {
		return WorkerPoolOperationAbortedError
	}

	ctx, cancel := context.WithCancel(ctx)
	wp.q = cancel

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

func (wp *WorkerPool) Shutdown() error {
	wp.Lock()
	defer wp.Unlock()

	if wp.status != WorkerPoolStatusRunning {
		return WorkerPoolOperationAbortedError
	}
	wp.status = WorkerPoolStatusStopping
	close(wp.tasks)
	wp.q()
	wp.wg.Wait()
	wp.status = WorkerPoolStatusStopped
	return nil
}

func New(capacity int) *WorkerPool {
	return &WorkerPool{
		wg:       sync.WaitGroup{},
		capacity: capacity,
		tasks:    make(chan Task),
		status:   WorkerPoolStatusStopped,
	}
}

type Task interface {
	Run(ctx context.Context) error
}
