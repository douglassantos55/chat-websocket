package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	currentId uint

	Sockets map[uint]*Socket
}

type PrivateMessage struct {
	Message  string `json:"message"`
	Receiver uint   `json:"receiver"`
}

func NewServer() *Server {
	return &Server{
		currentId: 1,

		Sockets: make(map[uint]*Socket),
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

		for {
			_, msg, err := c.ReadMessage()

			if err != nil {
				delete(s.Sockets, id)
				break
			}

			var data PrivateMessage
			jsonError := json.Unmarshal(msg, &data)

			if jsonError != nil {
				go s.broadcast(msg, c)
			} else {
				go s.private(data)
			}
		}
	}()
}

func (s *Server) broadcast(msg []byte, sender *websocket.Conn) {
	for _, socket := range s.Sockets {
		if socket.Conn != sender {
			socket.Conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func (s *Server) private(data PrivateMessage) {
	receiver, exists := s.Sockets[data.Receiver]

	if exists {
		receiver.Conn.WriteMessage(websocket.TextMessage, []byte(data.Message))
	}
}
