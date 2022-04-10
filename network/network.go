package network

import (
	"fmt"
	"net"
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
