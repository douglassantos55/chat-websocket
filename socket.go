package main

import "github.com/gorilla/websocket"

type Socket struct {
	Conn     *websocket.Conn
	Outgoing chan string
}

func NewSocket(socket *websocket.Conn) *Socket {
	return &Socket{
		Conn:     socket,
		Outgoing: make(chan string),
	}
}
