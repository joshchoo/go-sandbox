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
	loader := NewMultiRequestDataLoader(func(arg int64) (int64, error) {
		counter.Add(1)
		time.Sleep(1 * time.Second)
		return arg, nil
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
			res, err := loader.Load(ctx, int64Arg(i%10))
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

type LoaderArg[TArg any] interface {
	Arg() TArg
	Key() string
}

type int64Arg int64

func (a int64Arg) Arg() int64 {
	return int64(a)
}
func (a int64Arg) Key() string {
	return strconv.FormatInt(int64(a), 10)
}

/*
Possible improvements:
- Add ability to evict SingleRequestDataLoader that is infrequently used (e.g. LRU).
*/
type MultiRequestDataLoader[TArg any, TRes any] struct {
	reqKeyToLoader map[string]*SingleRequestDataLoader[TRes]
	mu             *sync.Mutex
	loader         func(arg TArg) (TRes, error)
	closed         bool
	wg             sync.WaitGroup
}

func NewMultiRequestDataLoader[TArg any, TRes any](loader func(arg TArg) (TRes, error)) *MultiRequestDataLoader[TArg, TRes] {
	return &MultiRequestDataLoader[TArg, TRes]{
		reqKeyToLoader: make(map[string]*SingleRequestDataLoader[TRes]),
		mu:             &sync.Mutex{},
		loader:         loader,
		closed:         false,
		wg:             sync.WaitGroup{},
	}
}

func (r *MultiRequestDataLoader[TArg, TRes]) Load(ctx context.Context, la LoaderArg[TArg]) (TRes, error) {
	key := la.Key()
	arg := la.Arg()

	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		var zero TRes
		return zero, errors.New("multi request data loader is closed")
	}
	_, ok := r.reqKeyToLoader[key]
	if !ok {
		r.reqKeyToLoader[key] = NewSingleRequestDataLoader(func(ctx context.Context) (TRes, error) {
			return r.loader(arg)
		})
	}
	r.mu.Unlock()

	resCh := make(chan Result[TRes])
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		var zero TRes
		return zero, errors.New("multi request data loader is closed")
	}
	loader := r.reqKeyToLoader[key]
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		res, err := loader.Load(ctx)
		resCh <- Result[TRes]{
			value: res,
			err:   err,
		}
	}()
	r.mu.Unlock()
	res := <-resCh
	return res.value, res.err
}

func (r *MultiRequestDataLoader[TArg, TRes]) Close() {
	fmt.Println("Closing multi request data loader...")
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()
	r.wg.Wait()
	fmt.Println("Multi request data loader closed!")
}

type SingleRequestDataLoader[TRes any] struct {
	requests chan Request[TRes]
	closeCh  chan struct{}
	done     *sync.WaitGroup
}

func (r *SingleRequestDataLoader[TRes]) Close() {
	close(r.closeCh)
	fmt.Println("Closing data loader...")
	r.done.Wait()
	fmt.Println("Data loader closed!")
}

func NewSingleRequestDataLoader[TRes any](loader func(ctx context.Context) (TRes, error)) *SingleRequestDataLoader[TRes] {
	requests := make(chan Request[TRes])
	closeCh := make(chan struct{})
	var done sync.WaitGroup

	initLoader(requests, loader, closeCh, &done)

	return &SingleRequestDataLoader[TRes]{
		requests: requests,
		closeCh:  closeCh,
		done:     &done,
	}
}

func initLoader[TRes any](requests chan Request[TRes], loader func(ctx context.Context) (TRes, error), closeCh chan struct{}, done *sync.WaitGroup) {
	done.Add(1)

	go func() {
		defer done.Done()
		resp := make(chan Result[TRes], 1)
		state := Idle
		var subscribers []chan Result[TRes]

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
						resp <- Result[TRes]{
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

func (r *SingleRequestDataLoader[TRes]) Load(ctx context.Context) (TRes, error) {
	subscriber := make(chan Result[TRes])
	req := Request[TRes]{
		ctx: ctx,
		out: subscriber,
	}
	select {
	case r.requests <- req:
	case <-r.closeCh:
		var zero TRes
		return zero, errors.New("data loader is closed")
	}
	result := <-subscriber
	if result.err != nil {
		var zero TRes
		return zero, result.err
	}
	return result.value, nil
}

type LoaderState uint8

const (
	Idle LoaderState = iota
	Loading
)

type Request[TRes any] struct {
	ctx context.Context
	out chan Result[TRes]
}

type Result[TRes any] struct {
	value TRes
	err   error
}
