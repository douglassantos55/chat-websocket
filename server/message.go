package server

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Id        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Sender    *Client                `json:"sender"`
}
