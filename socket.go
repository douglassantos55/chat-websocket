package main

import "github.com/gorilla/websocket"

type Socket struct {
	Conn *websocket.Conn
}

func NewSocket(socket *websocket.Conn) *Socket {
	return &Socket{
		Conn: socket,
	}
}
