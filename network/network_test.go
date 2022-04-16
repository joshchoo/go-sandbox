package network_test

import (
	"fmt"
	"io"
	"log"
	"net"
	"testing"

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
	shutdown := make(chan struct{})
	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				// Err from listener.Accept isn't necessarily fatal. It could be due to closing the listener connection.
				select {
				case <-shutdown:
					return
				default:
					t.Log(err)
					return
				}
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
	close(shutdown)
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
		n, err := conn.Read(b)
		if err != nil {
			if err != io.EOF {
				return err
			}
			// EOF is received when client closes the connection (i.e sends the FIN tcp packet)
			log.Println("[server] Received EOF.")
			return nil
		}
		log.Printf("[server] Read %q\n", b[:n])
	}
}
