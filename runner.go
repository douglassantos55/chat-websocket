package main

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
	receiverId := m.msg.Payload["receiver"].(uuid.UUID)
	receiver, exists := server.Clients[receiverId]

	if exists {
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

	log.Println("Broadcasting on channel", channelId)
	channel.Broadcast(b.msg)
}

type Channeler struct {
	msg Message
}

func (c *Channeler) Execute(server *Server) {
	channelId := c.msg.Payload["channel"].(uint)
	server.AddToChannel(c.msg.Sender, channelId)
}

type Authenticator struct {
	msg Message
}

func (a *Authenticator) Execute(server *Server) {
	a.msg.Sender.Name = a.msg.Payload["name"].(string)
	server.Clients[a.msg.Sender.Id] = a.msg.Sender

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
	case "join_channel":
		return &Channeler{msg}
	}

	return nil
}
