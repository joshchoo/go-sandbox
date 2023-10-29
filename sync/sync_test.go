package sync_test

import (
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"
)

func TestRWMutex(t *testing.T) {
	var sharedResource int
	var mu sync.RWMutex
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Info("Writer 1 start")
		mu.Lock()
		slog.Info("Writer 1 locked")
		time.Sleep(2 * time.Second)
		sharedResource = 111
		slog.Info("Writer 1 wrote")
		time.Sleep(1 * time.Second)
		slog.Info("Writer 1 unlocking")
		mu.Unlock()
		slog.Info("Writer 1 unlocked")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Info("Writer 2 start")
		mu.Lock()
		slog.Info("Writer 2 locked")
		time.Sleep(2 * time.Second)
		sharedResource = 222
		slog.Info("Writer 2 wrote")
		time.Sleep(2 * time.Second)
		slog.Info("Writer 2 unlocking")
		mu.Unlock()
		slog.Info("Writer 2 unlocked")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(50 * time.Millisecond)
		slog.Info("Reader 1 start")
		mu.RLock()
		slog.Info("Reader 1 read locked")
		slog.Info("Reader 1 read", "sharedResource", sharedResource)
		time.Sleep(1 * time.Second)
		slog.Info("Reader 1 unlocking")
		mu.RUnlock()
		slog.Info("Reader 1 unlocked")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)
		slog.Info("Reader 2 start")
		mu.RLock()
		slog.Info("Reader 2 read locked")
		slog.Info("Reader 2 read", "sharedResource", sharedResource)
		time.Sleep(1 * time.Second)
		slog.Info("Reader 2 unlocking")
		mu.RUnlock()
		slog.Info("Reader 2 unlocked")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(150 * time.Millisecond)
		slog.Info("Reader 3 start")
		mu.RLock()
		slog.Info("Reader 3 read locked")
		slog.Info("Reader 3 read", "sharedResource", sharedResource)
		time.Sleep(1 * time.Second)
		slog.Info("Reader 3 unlocking")
		mu.RUnlock()
		slog.Info("Reader 3 unlocked")
	}()

	wg.Wait()
}

// Readers cannot acquire a lock until there are no more writers?
// Multiple readers can hold a read lock at the same time as long as a write lock is not held.
// Is there some fairness mechanism for if a lock was Read Locked for too long compared to Write Locked?

// Suppose there is an existing Reader and pending Writer. Subsequent incoming Readers cannot acquire the lock until
// all existing Readers have released their locks. This is to prevent Writers waiting forever for all Readers to unlock.
// "a blocked (pending) Lock call excludes (prevents) new readers from acquiring the lock".
// In other words, Writers get higher priority than Readers?
// If there are many pending Writers, will that block pending Readers?
func TestAcquireReadThenWriteThenRead(t *testing.T) {
	var sharedResource int
	var mu sync.RWMutex
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Info("Reader 1 start")
		mu.RLock()
		slog.Info("Reader 1 read locked")
		slog.Info("Reader 1 read", "sharedResource", sharedResource)
		time.Sleep(2 * time.Second)
		slog.Info("Reader 1 unlocking")
		mu.RUnlock()
		slog.Info("Reader 1 unlocked")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(50 * time.Millisecond)
		slog.Info("Writer 1 start")
		mu.Lock()
		slog.Info("Writer 1 locked")
		time.Sleep(2 * time.Second)
		sharedResource = 111
		slog.Info("Writer 1 wrote")
		time.Sleep(1 * time.Second)
		slog.Info("Writer 1 unlocking")
		mu.Unlock()
		slog.Info("Writer 1 unlocked")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)
		slog.Info("Reader 2 start")
		mu.RLock()
		slog.Info("Reader 2 read locked")
		slog.Info("Reader 2 read", "sharedResource", sharedResource)
		time.Sleep(1 * time.Second)
		slog.Info("Reader 2 unlocking")
		mu.RUnlock()
		slog.Info("Reader 2 unlocked")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)
		slog.Info("Reader 3 start")
		mu.RLock()
		slog.Info("Reader 3 read locked")
		slog.Info("Reader 3 read", "sharedResource", sharedResource)
		time.Sleep(1 * time.Second)
		slog.Info("Reader 3 unlocking")
		mu.RUnlock()
		slog.Info("Reader 3 unlocked")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)
		slog.Info("Reader 4 start")
		mu.RLock()
		slog.Info("Reader 4 read locked")
		slog.Info("Reader 4 read", "sharedResource", sharedResource)
		time.Sleep(1 * time.Second)
		slog.Info("Reader 4 unlocking")
		mu.RUnlock()
		slog.Info("Reader 4 unlocked")
	}()

	wg.Wait()
}

func TestCopyMutexAfterFirstUse(t *testing.T) {
	var mu sync.Mutex
	go func() {
		mu.Lock()
		fmt.Println("Locked")
		time.Sleep(2 * time.Second)
		mu.Unlock()
		fmt.Println("Unlocked")
	}()

	time.Sleep(100 * time.Millisecond)
	muCopy := mu // copy a mutex, and expect this main goroutine to experience a deadlock. The other spawned goroutine will exit with no issues though.
	muCopy.Lock()
	fmt.Println("this line will never be reached even when mu is unlocked. This is because muCopy started with a non-zero counter, and mu.Unlock() does not change muCopy's internal state")
	muCopy.Unlock()
}
