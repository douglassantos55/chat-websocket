package main

import "log"

type Runner interface {
	Execute(server *Server)
}

type PrivMessenger struct {
	msg Message
}

func (m *PrivMessenger) Execute(server *Server) {
	receiver, exists := server.Sockets[m.msg.Receiver]

	if exists {
		receiver.Conn.WriteJSON(m.msg)
	}
}

type Broadcaster struct {
	msg Message
}

func (b *Broadcaster) Execute(server *Server) {
	channel := server.Channels[b.msg.Channel]

	if channel == nil {
		log.Printf("Channel %d not found", b.msg.Channel)
		return
	}

	sender := server.Sockets[b.msg.Sender]

	for _, socket := range channel.Sockets {
		if socket.Conn != sender.Conn {
			socket.Conn.WriteJSON(b.msg)
		}
	}
}

type Channeler struct {
	msg Message
}

func (c *Channeler) Execute(server *Server) {
	server.AddToChannel(c.msg.Socket, c.msg.Channel)
}

func NewMessageRunner(msg Message) Runner {
	switch msg.Type {
	case "broadcast":
		return &Broadcaster{msg}
	case "priv_msg":
		return &PrivMessenger{msg}
	case "join_channel":
		return &Channeler{msg}
	}

	return nil
}
