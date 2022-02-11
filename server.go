package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Channel struct {
	Id      uint
	Name    string
	Sockets map[uint]*Socket
}

type Server struct {
	mut       *sync.Mutex
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
		mut:       new(sync.Mutex),
		Sockets:   make(map[uint]*Socket),
		Channels:  map[uint]*Channel{1: defaultChannel},
	}
}

func (s *Server) Listen(addr string) {
	log.Printf("listening at %s", addr)
	http.HandleFunc("/", s.HandleConnection)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (s *Server) AddChannel(name string) uint {
	id := uint(len(s.Channels) + 1)

	s.Channels[id] = &Channel{
		Id:      id,
		Name:    name,
		Sockets: make(map[uint]*Socket),
	}

	return id
}

func (s *Server) AddToChannel(socket *Socket, channelId uint) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.Channels[channelId].Sockets[socket.Id] = socket
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

		id := s.currentId
		socket := NewSocket(c, id)

		s.mut.Lock()

		s.Sockets[id] = socket
		s.currentId += 1

		// add client to the default channel
		s.Channels[1].Sockets[id] = socket
		s.mut.Unlock()

		for {
			_, msg, err := c.ReadMessage()

			if err != nil {
				s.mut.Lock()
				delete(s.Sockets, id)

				for _, channel := range s.Channels {
					delete(channel.Sockets, id)
				}
				s.mut.Unlock()

				break
			}

			if msg == nil {
				continue
			}

			go s.parseMessage(msg, socket)
		}
	}()
}

func (s *Server) parseMessage(msg []byte, socket *Socket) {
	var message Message
	err := json.Unmarshal(msg, &message)

	message.Socket = socket

	if err != nil {
		return
	}

	if runner := NewMessageRunner(message); runner != nil {
		runner.Execute(s)
	}
}
