package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

var ErrInvalidNumberOfGoroutines = errors.New("invalid number of goroutines")

var ErrInvalidNumberOfMaxErrors = errors.New("invalid number of max errors")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrInvalidNumberOfGoroutines
	}

	if m <= 0 {
		return ErrInvalidNumberOfMaxErrors
	}

	chWork := make(chan Task, len(tasks))
	for i := range tasks {
		chWork <- tasks[i]
	}
	close(chWork)

	var errCount int64

	wg := sync.WaitGroup{}
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range chWork {
				if isDone(&errCount, m) {
					return
				}
				if err := task(); err != nil {
					atomic.AddInt64(&errCount, 1)
					if isDone(&errCount, m) {
						return
					}
				}
			}
		}()
	}

	wg.Wait()

	if isDone(&errCount, m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func isDone(count *int64, m int) bool {
	return atomic.LoadInt64(count) >= int64(m)
}
