package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	runMultiDataLoader()
	//runSimpleDataLoader()
}

func runSimpleDataLoader() {
	requestsCount := 25
	counter := atomic.Uint64{}
	loader := NewSingleRequestDataLoader(func(ctx context.Context) (int64, error) {
		counter.Add(1)
		time.Sleep(1 * time.Second)
		return rand.Int63(), nil
	})
	var wg sync.WaitGroup
	ctx := context.Background()

	wg.Add(requestsCount)
	go func() {
		time.Sleep(2 * time.Second)
		loader.Close()
	}()
	for i := 0; i < requestsCount; i++ {
		time.Sleep(100 * time.Millisecond)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("submitting request %d\n", i)
			res, err := loader.Load(ctx)
			if err != nil {
				fmt.Printf("request %d err: %v\n", i, err)
			} else {
				fmt.Printf("request %d: %v\n", i, res)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("Done!")
	fmt.Printf("Loader function called %d times.\n", counter.Load())
}

func runMultiDataLoader() {
	requestsCount := 200
	counter := atomic.Uint64{}
	loader := NewMultiRequestDataLoader(func(key string) (int64, error) {
		counter.Add(1)
		time.Sleep(1 * time.Second)
		return strconv.ParseInt(key, 10, 64)
	})
	var wg sync.WaitGroup
	ctx := context.Background()

	wg.Add(requestsCount)
	go func() {
		time.Sleep(2 * time.Second)
		loader.Close()
	}()
	for i := 0; i < requestsCount; i++ {
		time.Sleep(10 * time.Millisecond)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("submitting request %d\n", i)
			res, err := loader.Load(ctx, strconv.Itoa(i%10))
			if err != nil {
				fmt.Printf("request %d err: %v\n", i, err)
			} else {
				fmt.Printf("request %d: %v\n", i, res)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("Done!")
	fmt.Printf("Loader function called %d times.\n", counter.Load())
}

type MultiRequestDataLoader struct {
	reqKeyToLoader map[string]*SingleRequestDataLoader
	mu             *sync.Mutex
	loader         func(key string) (int64, error)
	closed         bool
	wg             sync.WaitGroup
}

func NewMultiRequestDataLoader(loader func(key string) (int64, error)) *MultiRequestDataLoader {
	return &MultiRequestDataLoader{
		reqKeyToLoader: make(map[string]*SingleRequestDataLoader),
		mu:             &sync.Mutex{},
		loader:         loader,
		closed:         false,
		wg:             sync.WaitGroup{},
	}
}

func (r *MultiRequestDataLoader) Load(ctx context.Context, key string) (int64, error) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return 0, errors.New("multi request data loader is closed")
	}
	_, ok := r.reqKeyToLoader[key]
	if !ok {
		r.reqKeyToLoader[key] = NewSingleRequestDataLoader(func(ctx context.Context) (int64, error) {
			return r.loader(key)
		})
	}
	r.mu.Unlock()

	resCh := make(chan Result)
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return 0, errors.New("multi request data loader is closed")
	}
	loader := r.reqKeyToLoader[key]
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		res, err := loader.Load(ctx)
		resCh <- Result{
			value: res,
			err:   err,
		}
	}()
	r.mu.Unlock()
	res := <-resCh
	return res.value, res.err
}

func (r *MultiRequestDataLoader) Close() {
	fmt.Println("Closing multi request data loader...")
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()
	r.wg.Wait()
	fmt.Println("Multi request data loader closed!")
}

type SingleRequestDataLoader struct {
	requests chan Request
	closeCh  chan struct{}
	done     *sync.WaitGroup
}

func (r *SingleRequestDataLoader) Close() {
	close(r.closeCh)
	fmt.Println("Closing data loader...")
	r.done.Wait()
	fmt.Println("Data loader closed!")
}

func NewSingleRequestDataLoader(loader func(ctx context.Context) (int64, error)) *SingleRequestDataLoader {
	requests := make(chan Request)
	closeCh := make(chan struct{})
	var done sync.WaitGroup

	initLoader(requests, loader, closeCh, &done)

	return &SingleRequestDataLoader{
		requests: requests,
		closeCh:  closeCh,
		done:     &done,
	}
}

func initLoader(requests chan Request, loader func(ctx context.Context) (int64, error), closeCh chan struct{}, done *sync.WaitGroup) {
	done.Add(1)

	type LoaderState uint8

	const (
		Idle LoaderState = iota
		Loading
	)

	go func() {
		defer done.Done()
		resp := make(chan Result, 1)
		state := Idle
		var subscribers []chan Result

		for {
			switch state {
			case Idle:
				select {
				case <-closeCh:
					return
				case req := <-requests:
					subscribers = append(subscribers, req.out)
					state = Loading
					go func() {
						value, err := loader(req.ctx)
						resp <- Result{
							value: value,
							err:   err,
						}
					}()
				}
			case Loading:
				select {
				case result := <-resp:
					for _, sub := range subscribers {
						sub <- result
					}
					// Reset
					subscribers = nil
					state = Idle
				case req := <-requests:
					subscribers = append(subscribers, req.out)
				}
			}
		}
	}()
}

type Request struct {
	ctx context.Context
	out chan Result
}

type Result struct {
	value int64
	err   error
}

func (r *SingleRequestDataLoader) Load(ctx context.Context) (int64, error) {
	subscriber := make(chan Result)
	req := Request{
		ctx: ctx,
		out: subscriber,
	}
	select {
	case r.requests <- req:
	case <-r.closeCh:
		return 0, errors.New("data loader is closed")
	}
	result := <-subscriber
	if result.err != nil {
		return 0, result.err
	}
	return result.value, nil
}
