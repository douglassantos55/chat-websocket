package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const DEFAULT_CHANNEL = 1

type Channel struct {
	Id      uint   `json:"id"`
	Name    string `json:"name"`
	clients map[uuid.UUID]*Client
	mut     *sync.Mutex
}

func NewChannel(id uint, name string) *Channel {
	return &Channel{
		Id:      id,
		Name:    name,
		clients: make(map[uuid.UUID]*Client),
		mut:     new(sync.Mutex),
	}
}

func (c *Channel) AddClient(client *Client) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.clients[client.Id] = client
}

func (c *Channel) RemoveClient(client *Client) {
	c.mut.Lock()
	defer c.mut.Unlock()
	delete(c.clients, client.Id)
}

func (c *Channel) Broadcast(msg Message) {
	for _, client := range c.clients {
		client.SendMessage(msg)
	}
}

func (c *Channel) Clients() int {
	return len(c.clients)
}

type Server struct {
	mut      *sync.Mutex
	Channels map[uint]*Channel
	Clients  map[uuid.UUID]*Client
}

func NewServer() *Server {
	defaultChannel := NewChannel(DEFAULT_CHANNEL, "Broadcast")

	return &Server{
		mut:      new(sync.Mutex),
		Clients:  make(map[uuid.UUID]*Client),
		Channels: map[uint]*Channel{DEFAULT_CHANNEL: defaultChannel},
	}
}

func (s *Server) Listen(addr string) {
	log.Printf("listening at %s", addr)
	http.HandleFunc("/", s.HandleConnection)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (s *Server) AddClient(client *Client) {
    s.mut.Lock()
    defer s.mut.Unlock()
	s.Clients[client.Id] = client
}

func (s *Server) RemoveClient(client *Client) {
    s.mut.Lock()
    defer s.mut.Unlock()
    delete(s.Clients, client.Id)
}

func (s *Server) AddChannel(name string) uint {
	s.mut.Lock()
	defer s.mut.Unlock()

	id := uint(len(s.Channels) + 1)

	s.Channels[id] = NewChannel(id, name)
	return id
}

func (s *Server) AddToChannel(client *Client, channelId uint) {
	s.Channels[channelId].AddClient(client)
}

func (s *Server) RemoveFromChannel(client *Client, channelId uint) {
	s.Channels[channelId].RemoveClient(client)
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

	client := NewClient(c)
    s.AddClient(client)

	go func() {
		defer client.Close()

		for {
			msg, err := client.GetMessage()

			if err != nil {
                s.RemoveClient(client)

				for _, channel := range s.Channels {
					channel.RemoveClient(client)
				}

				break
			}

			if runner := NewMessageRunner(msg); runner != nil {
				go runner.Execute(s)
			}
		}
	}()
}
