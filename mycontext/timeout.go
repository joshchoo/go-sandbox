package mycontext

import (
	"context"
	"fmt"
	"time"
)

func DoTaskWithTimeout[Result any](task func() Result, timeout time.Duration) (Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set channel size to 1 so that the goroutine can exit without causing a leak if the task timeout is exceeded.
	// See "Forgotten Sender": https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html
	resultCh := make(chan Result, 1)
	go func() {
		resultCh <- task()
	}()

	select {
	case <-ctx.Done():
		var zero Result
		return zero, fmt.Errorf("task canceled: %w", ctx.Err())
	case result := <-resultCh:
		return result, nil
	}
}
