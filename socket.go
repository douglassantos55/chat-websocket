package main

import "github.com/gorilla/websocket"

type Socket struct {
	Id   uint
	Conn *websocket.Conn
}

func NewSocket(socket *websocket.Conn, id uint) *Socket {
	return &Socket{
		Id:   id,
		Conn: socket,
	}
}
