package main

import "example.com/websocket/server"

func main() {
	server := server.NewServer()
	server.AddChannel("Games")
	server.AddChannel("Music")
	server.Listen("0.0.0.0:8080")
}
