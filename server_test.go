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

	c1.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, msg, err := c1.ReadMessage()

	if err == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", string(msg))
	}

	c2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, broadcast, err2 := c2.ReadMessage()

	if err2 != nil {
		t.Errorf("Expected message, got error: \"%s\"", err)
	}

	if string(broadcast) != "hello" {
		t.Errorf("Expected hello, got %s", string(broadcast))
	}
}

func TestPrivateMessage(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")
	c3 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()
	defer c3.Close()

	c1.WriteJSON(PrivateMessage{
		Message:  "hello, number 8",
		Receiver: 8,
        Sender: 7,
	})

	c2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

    var data PrivateMessage
    err := c2.ReadJSON(&data)

    if err != nil {
        t.Errorf("Expected message, got error: \"%s\"", err)
    }

    if data.Message != "hello, number 8" {
        t.Errorf("Expected hello, number 8, got %s", data.Message)
    }

    if data.Sender != 7 {
        t.Errorf("Expected sender 7, got %d", data.Sender)
    }

	c3.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
    _, message, err := c3.ReadMessage()

    if err == nil {
        t.Errorf("Expected timeout, got %s", string(message))
    }
}

func TestPrivateMessageToInvalid(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()

	c1.WriteJSON(PrivateMessage{
		Message:  "hello, number 8",
		Receiver: 999,
        Sender: 10,
	})

	c2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
    _, msg, err := c2.ReadMessage()

    if err == nil {
        t.Errorf("Expected timeout, got %s", string(msg))
    }
}
