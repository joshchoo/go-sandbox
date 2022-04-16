package network

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func AvailablePort() (int, func(), error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, nil, err
	}
	tcpAddress, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, nil, fmt.Errorf("Unable to cast to *net.TCPAddr: %q", listener.Addr())
	}
	close := func() {
		listener.Close()
	}
	return tcpAddress.Port, close, nil
}

const defaultPingInterval = 30 * time.Second

func Pinger(ctx context.Context, w io.Writer, resetTimerIntervalCh <-chan time.Duration) {
	interval := defaultPingInterval
	setInterval := func(i time.Duration) {
		switch {
		case i > 0:
			interval = i
		default:
			interval = defaultPingInterval
		}
	}

	timer := time.NewTimer(interval)
	stopTimerAndDrain := func() {
		if !timer.Stop() {
			<-timer.C
		}
	}
	defer stopTimerAndDrain()

	sendPing := func() (int, error) {
		return w.Write([]byte("ping"))
	}

	// The following things can happen during the loop:
	// 1. Pinger is cancelled.
	// 2. Timer interval is changed.
	// 3. Timer expires.
	for {
		select {
		case <-ctx.Done():
			return
		case newInterval := <-resetTimerIntervalCh:
			stopTimerAndDrain()
			setInterval(newInterval)
			log.Printf("[Pinger] Updated interval to %v\n", interval)
		case <-timer.C: // send "ping" when timer expires
			log.Println("[Pinger] Sending ping")
			if _, err := sendPing(); err != nil {
				return
			}
		}

		// Start the timer again
		timer.Reset(interval)
	}
}
