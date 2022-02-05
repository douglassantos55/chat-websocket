package main

import "github.com/gorilla/websocket"

type Socket struct {
	Conn     *websocket.Conn
	Incoming chan string
}

func NewSocket(socket *websocket.Conn) *Socket {
	return &Socket{
		Conn:     socket,
		Incoming: make(chan string),
	}
}
