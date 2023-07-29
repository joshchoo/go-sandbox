package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	runSimpleDataLoader()
}

func runSimpleDataLoader() {
	requestsCount := 25
	var loader DataLoader = NewSingleRequestDataLoader(getRandomNumberDelayed)
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
}

type DataLoader interface {
	Load(ctx context.Context) (int64, error)
	Close()
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

func NewSingleRequestDataLoader(loaderFn func() (int64, error)) *SingleRequestDataLoader {
	requests := make(chan Request)
	closeCh := make(chan struct{})
	var done sync.WaitGroup

	initLoader(requests, loaderFn, closeCh, &done)

	return &SingleRequestDataLoader{
		requests: requests,
		closeCh:  closeCh,
		done:     &done,
	}
}

func initLoader(requests chan Request, loaderFn func() (int64, error), closeCh chan struct{}, done *sync.WaitGroup) {
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
						value, err := loaderFn()
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

func getRandomNumberDelayed() (int64, error) {
	time.Sleep(1 * time.Second)
	return rand.Int63(), nil
}
