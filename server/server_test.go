package server

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

func Broadcast(message string, channel uint) Message {
	return Message{
		Type: "broadcast",
		Payload: map[string]interface{}{
			"message": message,
			"channel": channel,
		},
	}
}

func SendPrivate(message string, userId string) Message {
	return Message{
		Type: "priv_msg",
		Payload: map[string]interface{}{
			"receiver": userId,
			"message":  message,
		},
	}
}

func JoinChannel(channelId uint) Message {
	return Message{
		Type: "join_channel",
		Payload: map[string]interface{}{
			"channel": channelId,
		},
	}
}

func LeaveChannel(channelId uint) Message {
	return Message{
		Type: "leave_channel",
		Payload: map[string]interface{}{
			"channel": channelId,
		},
	}
}

func Authenticate(name string) Message {
	return Message{
		Type: "auth",
		Payload: map[string]interface{}{
			"name": name,
		},
	}
}

func TestAcceptsConnections(t *testing.T) {
	go server.Listen("0.0.0.0:8080")

	// wait a bit, don't know how to do this right
	time.Sleep(100 * time.Millisecond)

	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()

	if len(server.Clients) != 2 {
		t.Errorf("Expected 2 connection, got %d", len(server.Clients))
	}
}

func TestClosesConnections(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	c1.Close()
	defer c2.Close()

	// wait a bit, don't know how to do this right
	time.Sleep(100 * time.Millisecond)

	if len(server.Clients) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(server.Clients))
	}
}

func TestAuthentication(t *testing.T) {
    server.AddChannel("Games")

	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()

	c1.WriteJSON(Authenticate("john doe"))

	// wait a bit, don't know how to do this right
	time.Sleep(100 * time.Millisecond)

	for _, channel := range server.Channels {
		if channel.Clients() != 1 {
			t.Errorf("Expected 1 connection on all channels, got %d in channel %s", channel.Clients(), channel.Name)
		}
	}
}

func TestBroadcastMessages(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()

	c1.WriteJSON(Authenticate("john doe"))
	c2.WriteJSON(Authenticate("jane doe"))

	// wait a bit, don't know how to do this right
	time.Sleep(100 * time.Millisecond)

	var msg Message

	// Skip auth returns
	c1.ReadJSON(&msg)
	c2.ReadJSON(&msg)

	c1.WriteJSON(Broadcast("hello", 1))

	c1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	err := c1.ReadJSON(&msg)

	if err != nil {
		t.Errorf("Expected message, got error: \"%s\"", err)
	}

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	err2 := c2.ReadJSON(&msg)

	if err2 != nil {
		t.Errorf("Expected message, got error: \"%s\"", err2)
	}

	if msg.Payload["message"].(string) != "hello" {
		t.Errorf("Expected hello, got %s", msg.Payload["message"])
	}
}

func TestRemoveSocketFromChannels(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	c2.Close()
	defer c1.Close()

	c1.WriteJSON(Authenticate("john doe"))
	c2.WriteJSON(Authenticate("jane doe"))

	time.Sleep(100 * time.Millisecond)

	if len(server.Clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(server.Clients))
	}

	for _, channel := range server.Channels {
		if channel.Clients() != 1 {
            t.Errorf("Expected 1 client on all channels, got %d on channel %s", channel.Clients(), channel.Name)
		}
	}
}

func TestJoinChannel(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()

	channelId := server.AddChannel("Games")
	c1.WriteJSON(JoinChannel(channelId))

	time.Sleep(100 * time.Millisecond)

	if server.Channels[channelId].Clients() != 1 {
		t.Errorf("Expected 1 connection, got %d", server.Channels[channelId].Clients())
	}
}

func TestLeaveChannel(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()

    c1.WriteJSON(Authenticate("john doe"))
    c2.WriteJSON(Authenticate("jane doe"))

    var msg Message

	// Skip auth returns
	c1.ReadJSON(&msg)
	c2.ReadJSON(&msg)

    time.Sleep(100 * time.Millisecond)

    c2.WriteJSON(LeaveChannel(DEFAULT_CHANNEL))

    time.Sleep(100 * time.Millisecond)

    c1.WriteJSON(Broadcast("hi from leave channel", DEFAULT_CHANNEL))

    if server.Channels[DEFAULT_CHANNEL].Clients() != 1 {
        t.Errorf("Expected 1 client, got %d", server.Channels[DEFAULT_CHANNEL].Clients())
    }

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

    err := c2.ReadJSON(&msg)

    if err == nil {
        t.Errorf("Expected timeout, got message %s", msg.Payload["message"])
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

	c1.WriteJSON(JoinChannel(channelId))
	c2.WriteJSON(JoinChannel(channelId))

	time.Sleep(100 * time.Millisecond)

	c1.WriteJSON(Broadcast("hi people from channel", channelId))

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var msg Message
	err := c2.ReadJSON(&msg)

	if err != nil {
		t.Errorf("Expected message, got error: \"%s\"", err)
	} else {
		if msg.Payload["message"] != "hi people from channel" {
			t.Errorf("Expected \"hi people from channel\", got %s", msg.Payload["message"])
		}
	}

	c3.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var fail Message
	timeout := c3.ReadJSON(&fail)

	if timeout == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", fail.Payload["message"])
	}
}

func TestReceivesFromAllJoinedChannels(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")
	c3 := ConnectToServer("0.0.0.0:8080")

	defer c1.Close()
	defer c2.Close()
	defer c3.Close()

	c1.WriteJSON(JoinChannel(DEFAULT_CHANNEL))
	c2.WriteJSON(JoinChannel(DEFAULT_CHANNEL))
	c3.WriteJSON(JoinChannel(DEFAULT_CHANNEL))

	channelId := server.AddChannel("Jobs")

	c1.WriteJSON(JoinChannel(channelId))
	c2.WriteJSON(JoinChannel(channelId))

	c1.WriteJSON(Broadcast("hi people from channel", channelId))

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var msg Message
	err := c2.ReadJSON(&msg)

	if err != nil {
		t.Errorf("Expected message, got error: \"%s\"", err)
	} else {
		if msg.Payload["message"] != "hi people from channel" {
			t.Errorf("Expected \"hi people from channel\", got %s", msg.Payload["message"])
		}
	}

	c3.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var fail Message
	timeout := c3.ReadJSON(&fail)

	if timeout == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", fail.Payload["message"])
	}

	c1.WriteJSON(Broadcast("hi people from broadcast", DEFAULT_CHANNEL))

	c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	broadcastErr := c2.ReadJSON(&msg)

	if broadcastErr != nil {
		t.Errorf("Expected message, got error: \"%s\"", broadcastErr)
	} else {
		if msg.Payload["message"] != "hi people from broadcast" {
			t.Errorf("Expected \"hi people from broadcast\", got %s", msg.Payload["message"])
		}
	}

	c3.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	timeout2 := c3.ReadJSON(&fail)

	if timeout2 == nil {
		t.Errorf("Expected timeout, got message: \"%s\"", fail.Payload["message"])
	}
}

func TestPrivateMessage(t *testing.T) {
	c1 := ConnectToServer("0.0.0.0:8080")
	c2 := ConnectToServer("0.0.0.0:8080")

	c1.WriteJSON(Authenticate("john doe"))
	c2.WriteJSON(Authenticate("jane doe"))

	var msg Message
	err := c2.ReadJSON(&msg)

	if err != nil {
		t.Errorf("Expected auth response, got error %s", err)
	}

	receiver := msg.Payload["user"].(map[string]interface{})
	c1.WriteJSON(SendPrivate("hi user", receiver["id"].(string)))

	privErr := c2.ReadJSON(&msg)

	if privErr != nil {
		t.Errorf("Expected message, got %s", privErr)
	}

	if msg.Payload["message"].(string) != "hi user" {
		t.Errorf("Expected \"hi user\", got %s", msg.Payload["message"].(string))
	}
}
