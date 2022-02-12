package server

import "time"

type Message struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Sender    *Client                `json:"sender"`
}
