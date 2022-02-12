package server

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	Id     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	socket *websocket.Conn
}

func NewClient(socket *websocket.Conn) *Client {
	return &Client{
		Id:     uuid.New(),
		socket: socket,
	}
}

func (c *Client) Close() {
	c.socket.Close()
}

func (c *Client) GetMessage() (Message, error) {
	var message Message
	err := c.socket.ReadJSON(&message)

	if err != nil {
		return Message{}, err
	}

	message.Sender = c
	return message, nil
}

func (c *Client) SendMessage(message Message) error {
	message.Timestamp = time.Now()
	return c.socket.WriteJSON(message)
}
