package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Channel struct {
	Id      uint
	Name    string
	Sockets map[uint]*Socket
}

type Server struct {
	currentId uint
	Channels  map[uint]*Channel
	Sockets   map[uint]*Socket
}

func NewServer() *Server {
	defaultChannel := &Channel{
		Id:      1,
		Name:    "Broadcast",
		Sockets: make(map[uint]*Socket),
	}

	return &Server{
		currentId: 1,
		Sockets:   make(map[uint]*Socket),
		Channels:  map[uint]*Channel{1: defaultChannel},
	}
}

func (s *Server) Listen(addr string) {
	log.Printf("listening at %s", addr)
	http.HandleFunc("/", s.HandleConnection)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	go func() {
		defer c.Close()

		socket := NewSocket(c)
		id := s.currentId
		s.Sockets[id] = socket
		s.currentId += 1

		// add client to the default channel
		s.Channels[1].Sockets[id] = socket

		for {
			_, msg, err := c.ReadMessage()

			if err != nil {
				delete(s.Sockets, id)

				for _, channel := range s.Channels {
					delete(channel.Sockets, id)
				}

				break
			}

            if msg == nil {
                continue;
            }

			go func() {
				var message Message
				err := json.Unmarshal(msg, &message)

				if err == nil {
					go s.broadcast(message, c)
				}
			}()

			go func() {
				var message PrivateMessage
				err := json.Unmarshal(msg, &message)

				if err == nil {
					go s.private(message, id)
				}
			}()
		}
	}()
}

func (s *Server) broadcast(message Message, sender *websocket.Conn) {
	channel := s.Channels[message.Channel]

	if channel == nil {
		log.Printf("Channel %d not found", message.Channel)
		return
	}

	for _, socket := range channel.Sockets {
		if socket.Conn != sender {
			socket.Conn.WriteJSON(message)
		}
	}
}

func (s *Server) private(data PrivateMessage, senderId uint) {
	receiver, exists := s.Sockets[data.Receiver]

	if exists {
		data.Sender = senderId
		receiver.Conn.WriteJSON(data)
	}
}
