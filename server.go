package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	currentId uint
	Sockets   map[uint]*Socket
}

func NewServer() *Server {
	return &Server{
		currentId: 1,
		Sockets:   make(map[uint]*Socket),
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
			_, _, err := c.ReadMessage()

			if err != nil {
				delete(s.Sockets, id)
				break
			}
		}
	}()
}
