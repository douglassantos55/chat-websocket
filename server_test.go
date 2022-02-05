package main

import (
	"log"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var server = NewServer()

func ConnectToServer(addr string) *websocket.Conn {
	c, _, err := websocket.DefaultDialer.Dial("ws://"+addr, nil)

	if err != nil {
		log.Fatal("dial:", err)
	}

	return c
}

func TestAcceptsConnections(t *testing.T) {
    go server.Listen("0.0.0.0:8080")

	// wait a bit, don't know how to do this right
	time.Sleep(200 * time.Millisecond)

	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	if len(server.Sockets) != 2 {
		t.Errorf("Expected 2 connection, got %d", len(server.Sockets))
	}

	defer c1.Close()
	defer c2.Close()
}

func TestClosesConnections(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	c1.Close()

	// wait a bit, don't know how to do this right
	time.Sleep(200 * time.Millisecond)

	if len(server.Sockets) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(server.Sockets))
	}

	defer c2.Close()
}
