package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	echo "github.com/tkivisik/tcp-relay-server/sampleappecho"
)

// TestListenForConnections tests for network and port
func TestListenForConnections(t *testing.T) {
	tables := []struct {
		port int
	}{
		{8080},
		{9090},
	}
	for _, table := range tables {
		listener := listenForConnections(table.port)
		defer listener.Close()

		if address := listener.Addr().String(); strings.HasSuffix(address, strconv.Itoa(table.port)) == false {
			t.Errorf("Port not as expected, got: %s, want: %s.", address, strconv.Itoa(table.port))
		}
	}
}

// TestListenForConnectionsPanic tests for panic if port in use
func TestListenForConnectionsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	listener1 := listenForConnections(9990)
	listener2 := listenForConnections(9990)
	listener1.Close()
	listener2.Close()

}

// TestAcceptConnections tests for multiple connections per listener
// Currently the worst test :)
func TestAcceptConnections(t *testing.T) {
	wg := sync.WaitGroup{}
	listener := listenForConnections(31331)
	defer listener.Close()
	go func() {
		conn := acceptConnections(listener)
		defer conn.Close()
		fmt.Println(conn.RemoteAddr())
	}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			conn, err := net.Dial("tcp", ":31331")
			defer conn.Close()
			if err != nil {
				t.Errorf("Was expecting no errors in accepting multiple dials. %s", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// TestEcho makes a system level end to end test.
func TestEcho(t *testing.T) {
	message := []byte("testing123")
	regPort := 8080

	go RunTCPServer(regPort, 1)
	go echo.Run()
	// Give some time for the setup
	time.Sleep(time.Second / 200)

	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: regPort + 1,
		Zone: "",
	})
	if err != nil {
		t.Errorf("Connection failed, was expecting success on port %d. %s", regPort+1, err)
	}
	defer conn.Close()
	conn.Write(message)
	buf := make([]byte, 10, 10)
	conn.Read(buf)
	for i, char := range message {
		if char == buf[i] {
			continue
		}
		t.Errorf("Echo is not working, got: %s, want: %s", buf, message)
	}
}
