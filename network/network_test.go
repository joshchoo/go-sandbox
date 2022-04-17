package network_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/joshchoo/go-sandbox/network"
)

func TestFindAvailablePort(t *testing.T) {
	port, close, err := network.AvailablePort()
	if err != nil {
		t.Fatal(err)
	}
	defer close()
	fmt.Printf("Got available port: %d\n", port)
}

func TestDial(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("[server] Listening on %q\n", listener.Addr().String())

	done := make(chan struct{})
	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					log.Println("Network connection already closed. Returning.")
					return
				}
				t.Fatal(err)
			}
			go func() {
				if err := handleConn(conn, done); err != nil {
					t.Fatal(err)
				}
			}()
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("[client] Dialed: %q", conn.RemoteAddr())

	conn.Write([]byte("ping"))

	log.Println("[client] Closing dial connection")
	conn.Close()
	<-done
	listener.Close()
	<-done
}

func handleConn(conn net.Conn, done chan struct{}) error {
	defer func() {
		log.Println("[server] Closing connection.")
		conn.Close()
		done <- struct{}{}
	}()
	log.Println("[server] Handling new connection")

	for {
		b := make([]byte, 1024)
		// It is a good practice to set a deadline for Read/Write from the connection.
		if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
			return err
		}
		n, err := conn.Read(b)
		if err != nil {
			if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				continue
			} else if err != io.EOF {
				return err
			}
			// EOF is received when client closes the connection (i.e sends the FIN tcp packet)
			log.Println("[server] Received EOF.")
			return nil
		}
		log.Printf("[server] Read %q\n", b[:n])
	}
}

func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// net.Dial() actually creates a nil `net.Dialer` under-the-hood.
	d := net.Dialer{
		Control: func(_, addr string, _ syscall.RawConn) error {
			// Mock a DNS time-out error so that DialTimeout successfully makes a connection, but doesn't actually dials `address`.
			return &net.DNSError{
				Err:    "connection timed out",
				Name:   addr,
				Server: "127.0.0.1",
				// Specifying `IsTimeout: true` allows us to check that the `net.Error.Timeout()` is `true`
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
		// Timeout isn't actually used because we mocked `Control`
		Timeout: timeout,
	}
	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	// 10.0.0.1 is a non-routable address
	conn, err := DialTimeout("tcp", "10.0.0.1:http", 3*time.Second)

	if err == nil {
		conn.Close()
		t.Fatal("Expected connection to timeout, but it did not")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}
	if !nErr.Timeout() {
		t.Fatalf("Expected timeout error, but got %q", nErr)
	}
}

func TestDialContextDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	d := net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			// Simulate delay that is longer than the timeout.
			// The test will wait for this delay to complete, even if the timeout is shorter,
			// but the test will still fail with context.DeadlineExceeded.
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	}
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:http")

	if err == nil {
		conn.Close()
		t.Fatal("Expected connection to timeout, but it did not")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}
	// If context deadline is exceeded, net.Error.Timeout() will be true.
	if !nErr.Timeout() {
		t.Fatalf("Expected timeout error, but got %q", nErr)
	}
	// Ensure that the Err is due to deadline exceeded rather than context being cancelled.
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected deadline exceeded, got %q", ctx.Err())
	}

}

func TestDialContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	d := net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	}
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:http")

	cancel()

	if err == nil {
		conn.Close()
		t.Fatal("Expected connection to be canceled, but it did not")
	}
	if ctx.Err() != context.Canceled {
		t.Errorf("Expected canceled, got %q", ctx.Err())
	}
}

func TestPingerAdvanceDeadline(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	startTime := time.Now()
	done := make(chan struct{})

	// This goroutines simulates a server periodically pinging the client and checking for a heartbeat ("pong").
	// The server expects to receive a "pong" from the client at least once every five seconds to confirm that the client is still alive.
	go func() {
		defer func() { close(done) }()

		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Network connection already closed. Returning.")
				return
			}
			t.Fatal(err)
		}
		defer conn.Close()

		// Send a ping every second
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
		}()
		resetTimerIntervalCh := make(chan time.Duration, 1)
		resetTimerIntervalCh <- 1 * time.Second
		// Ping the client at each time interval
		go network.Pinger(ctx, conn, resetTimerIntervalCh)

		// Continuously read from connection
		buf := make([]byte, 1024)
		for {
			// Update the connection's read and write deadlines
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}

			n, err := conn.Read(buf)
			// error may occur due to connection timeout, disconnect, io.EOF, etc
			if err != nil {
				if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
					t.Logf("[server: %s] Read deadline timeout exceeded. Exiting.", time.Since(startTime).Truncate(time.Second))
					return
				}
				t.Error(err)
			}
			t.Logf("[server: %s] Got: %s", time.Since(startTime).Truncate(time.Second), buf[:n])

			// now reset the Pinger's timer to the default interval (30 seconds) so that we can trigger a read deadline exceeded.
			resetTimerIntervalCh <- 0
		}
	}()

	// Client dials the server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Receive 4 pings from the server. One ping per second.
	// The total wait time of 4 seconds is still within the server's read deadline of 5 seconds.
	// Then send a "pong" to the server.
	buf := make([]byte, 1024)
	for i := 0; i < 4; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[client: %s] Got: %s", time.Since(startTime).Truncate(time.Second), buf[:n])
	}
	_, err = conn.Write([]byte("pong"))
	if err != nil {
		t.Fatal(err)
	}

	// After the server received the "pong", the Ping interval changes to 30 seconds.
	// The server's read deadline will exceed before the first loop even completes,
	// and the client receives io.EOF.
	for i := 0; i < 4; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Fatal(err)
			}
			t.Logf("[client] Received EOF")
			break
		}
		t.Logf("[client: %s] Got: %s", time.Since(startTime).Truncate(time.Second), buf[:n])
	}

	// The server's read deadline will exceed 5 seconds, and return an error on conn.Read().
	// Then it will close the `done` channel. Wait for the channel to close.
	<-done

	end := time.Since(startTime).Truncate(time.Second)
	t.Logf("[%s] done", end)
	if end != 9*time.Second {
		t.Fatalf("Expected EOF at 9 seconds, got %s", end)
	}
}
