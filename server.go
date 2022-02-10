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

			go s.parseMessage(msg, id, socket)
		}
	}()
}

func (s *Server) parseMessage(msg []byte, id uint, socket *Socket) {
	message := s.parseBroadcast(msg)
	private := s.parsePrivate(msg)
	channel := s.parseChannel(msg)

	select {
	case m := <-message:
		go s.broadcast(m)
	case m := <-private:
		go s.private(m)
	case m := <-channel:
		s.mut.Lock()
		s.Channels[m.Channel].Sockets[id] = socket
		s.mut.Unlock()
	}

}

func (s *Server) parseBroadcast(msg []byte) chan Message {
	ch := make(chan Message)

	go func() {
		var message Message
		err := json.Unmarshal(msg, &message)

		if err == nil {
			ch <- message
		}
	}()

	return ch
}

func (s *Server) parsePrivate(msg []byte) chan PrivateMessage {
	ch := make(chan PrivateMessage)

	go func() {
		var message PrivateMessage
		err := json.Unmarshal(msg, &message)

		if err == nil {
			ch <- message
		}
	}()

	return ch
}

func (s *Server) parseChannel(msg []byte) chan JoinChannel {
	ch := make(chan JoinChannel)
	go func() {
		var message JoinChannel
		err := json.Unmarshal(msg, &message)

		if err == nil {
			ch <- message
		}
	}()
	return ch
}

func (s *Server) broadcast(message Message) {
	channel := s.Channels[message.Channel]

	if channel == nil {
		log.Printf("Channel %d not found", message.Channel)
		return
	}

	sender := s.Sockets[message.Sender]

	for _, socket := range channel.Sockets {
		if socket.Conn != sender.Conn {
			socket.Conn.WriteJSON(message)
		}
	}
}

func (s *Server) private(message PrivateMessage) {
	receiver, exists := s.Sockets[message.Receiver]

	if exists {
		receiver.Conn.WriteJSON(message)
	}
}
