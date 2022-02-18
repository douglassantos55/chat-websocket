package server

import (
	"log"

	"github.com/google/uuid"
)

type Runner interface {
	Execute(server *Server)
}

type PrivMessenger struct {
	msg Message
}

func (m *PrivMessenger) Execute(server *Server) {
	receiverId, _ := uuid.Parse(m.msg.Payload["receiver"].(string))
	receiver, exists := server.Clients[receiverId]

	if exists {
        m.msg.Payload["channel"] = receiver
        m.msg.Sender.SendMessage(m.msg)

        m.msg.Payload["channel"] = m.msg.Sender
		err := receiver.SendMessage(m.msg)

		if err != nil {
			log.Println("Could not send message: ", m.msg)
		}
	}
}

type Broadcaster struct {
	msg Message
}

func (b *Broadcaster) Execute(server *Server) {
	channelId := uint(b.msg.Payload["channel"].(float64))
	channel := server.Channels[channelId]

	if channel == nil {
		log.Printf("Channel %d not found", channelId)
		return
	}

	channel.Broadcast(b.msg)
}

type Channeler struct {
	msg Message
}

func (c *Channeler) Execute(server *Server) {
    if c.msg.Payload["channel"] == nil {
        return
    }

	channelId := uint(c.msg.Payload["channel"].(float64))

	switch c.msg.Type {
	case "join_channel":
		server.AddToChannel(c.msg.Sender, channelId)
	case "leave_channel":
		server.RemoveFromChannel(c.msg.Sender, channelId)
	}
}

type Authenticator struct {
	msg Message
}

func (a *Authenticator) Execute(server *Server) {
	a.msg.Sender.Name = a.msg.Payload["name"].(string)

    for _, channel := range server.Channels {
        channel.AddClient(a.msg.Sender)
    }

	a.msg.Sender.SendMessage(Message{
		Type: "auth",
		Payload: map[string]interface{}{
			"user":     a.msg.Sender,
			"channels": server.Channels,
		},
	})
}

func NewMessageRunner(msg Message) Runner {
	switch msg.Type {
	case "auth":
		return &Authenticator{msg}
	case "broadcast":
		return &Broadcaster{msg}
	case "priv_msg":
		return &PrivMessenger{msg}
	case "join_channel", "leave_channel":
		return &Channeler{msg}
	}

	return nil
}
