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

	defer c1.Close()
	defer c2.Close()

	if len(server.Sockets) != 2 {
		t.Errorf("Expected 2 connection, got %d", len(server.Sockets))
	}
}

func TestClosesConnections(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	c1.Close()
	defer c2.Close()

	// wait a bit, don't know how to do this right
	time.Sleep(200 * time.Millisecond)

	if len(server.Sockets) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(server.Sockets))
	}
}

func TestBroadcastMessages(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()

	c1.WriteMessage(websocket.TextMessage, []byte("hello"))

	c1.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := c1.ReadMessage()

	if err == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", string(msg))
	}

	c2.SetReadDeadline(time.Now().Add(time.Second))
	_, broadcast, err2 := c2.ReadMessage()

	if err2 != nil {
		t.Errorf("Expected message, got error: \"%s\"", err)
	}

	if string(broadcast) != "hello" {
		t.Errorf("Expected hello, got %s", string(broadcast))
	}
}
