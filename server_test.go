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

	c1.WriteJSON(Message{
		Message: "hello",
		Channel: 1,
		Sender:  5,
	})

	c1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var msg Message
	err := c1.ReadJSON(&msg)

	if err == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", msg.Message)
	}

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	err2 := c2.ReadJSON(&msg)

	if err2 != nil {
		t.Errorf("Expected message, got error: \"%s\"", err2)
	}

	if msg.Message != "hello" {
		t.Errorf("Expected hello, got %s", msg.Message)
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
		Sender:   7,
		Receiver: 8,
	})

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

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

	c3.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
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
		Sender:   10,
		Receiver: 999,
	})

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var msg PrivateMessage
	err := c2.ReadJSON(&msg)

	if err == nil {
		t.Errorf("Expected timeout, got %s", msg.Message)
	}
}

func TestRemoveSocketFromChannels(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	c2.Close()
	defer c1.Close()

	time.Sleep(100 * time.Millisecond)

	if socket := server.Channels[1].Sockets[13]; socket != nil {
		t.Error("Expected error, got socket", socket)
	}
}

func TestJoinChannel(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()

	channelId := server.AddChannel("Games")

	c1.WriteJSON(JoinChannel{
		Channel: channelId,
	})

	time.Sleep(100 * time.Millisecond)

	if len(server.Channels[2].Sockets) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(server.Channels[2].Sockets))
	}
}

func TestOnlyMembersOfChannelReceiveBroadcast(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")
	c3 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()
	defer c3.Close()

	channelId := server.AddChannel("Music")

	c1.WriteJSON(JoinChannel{
		Channel: channelId,
	})

	c2.WriteJSON(JoinChannel{
		Channel: channelId,
	})

	time.Sleep(100 * time.Millisecond)

	c1.WriteJSON(Message{
		Message: "hi people from channel",
		Channel: channelId,
		Sender:  16,
	})

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	c3.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var msg Message
	err := c2.ReadJSON(&msg)

	if err != nil {
		t.Errorf("Expected message, got error: \"%s\"", err)
	} else {
		if msg.Sender != 16 {
			t.Errorf("Expected sender to be 16, got %d", msg.Sender)
		}

		if msg.Message != "hi people from channel" {
			t.Errorf("Expected \"hi people from channel\", got %s", msg.Message)
		}
	}

	var fail Message
	timeout := c3.ReadJSON(&fail)

	if timeout == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", fail.Message)
	}
}

func TestReceivesFromAllJoinedChannels(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")
	c3 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()
	defer c3.Close()

	channelId := server.AddChannel("Jobs")

	c1.WriteJSON(JoinChannel{
		Channel: channelId,
	})

	c2.WriteJSON(JoinChannel{
		Channel: channelId,
	})

	c1.WriteJSON(Message{
		Message: "hi people from channel",
		Channel: channelId,
		Sender:  19,
	})

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var msg Message
	err := c2.ReadJSON(&msg)

	if err != nil {
		t.Errorf("Expected message, got error: \"%s\"", err)
	} else {
		if msg.Sender != 19 {
			t.Errorf("Expected sender to be 19, got %d", msg.Sender)
		}

		if msg.Message != "hi people from channel" {
			t.Errorf("Expected \"hi people from channel\", got %s", msg.Message)
		}
	}

	c3.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var fail Message
	timeout := c3.ReadJSON(&fail)

	if timeout == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", fail.Message)
	}

	c1.WriteJSON(Message{
		Message: "hi people from broadcast",
		Channel: 1,
		Sender:  19,
	})

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	broadcastErr := c2.ReadJSON(&msg)

	if broadcastErr != nil {
		t.Errorf("Expected message, got error: \"%s\"", broadcastErr)
	} else {
		if msg.Sender != 19 {
			t.Errorf("Expected sender to be 19, got %d", msg.Sender)
		}

		if msg.Message != "hi people from broadcast" {
			t.Errorf("Expected \"hi people from broadcast\", got %s", msg.Message)
		}
	}

	c3.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	timeout2 := c3.ReadJSON(&fail)

	if timeout2 == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", fail.Message)
	}
}
