package network_test

import (
	"fmt"
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
