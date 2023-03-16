package util

import "sync"

type Task func()

type WorkerPool struct {
	taskQueue chan Task
	wg        sync.WaitGroup
}

func NewWorkerPool(numWorkers int, queueSize int) *WorkerPool {
	pool := &WorkerPool{
		taskQueue: make(chan Task, queueSize),
	}

	for i := 0; i < numWorkers; i++ {
		go pool.worker()
	}

	return pool
}

func (pool *WorkerPool) worker() {
	for task := range pool.taskQueue {
		task()
	}
}

func (pool *WorkerPool) Submit(task Task) {
	pool.wg.Add(1)
	pool.taskQueue <- func() {
		defer pool.wg.Done()
		task()
	}
}

func (pool *WorkerPool) Wait() {
	pool.wg.Wait()
}

func (pool *WorkerPool) Stop() {
	close(pool.taskQueue)
}
