package main

import (
	"encoding/json"
	"errors"
)

type Message struct {
	Message string `json:"message"`
	Sender  uint   `json:"sender"`
	Channel uint   `json:"channel"`
}

func (m *Message) UnmarshalJSON(b []byte) error {
	var data map[string]interface{}

	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	message := data["message"]
	sender := data["sender"]
	channel := data["channel"]

	if message == nil || message.(string) == "" {
		return errors.New("No message")
	}
	if sender == nil || sender.(float64) == 0 {
		return errors.New("No sender")
	}
	if channel == nil || channel.(float64) == 0 {
		return errors.New("No channel")
	}

	m.Message = message.(string)
	m.Sender = uint(sender.(float64))
	m.Channel = uint(channel.(float64))

	return nil
}

type PrivateMessage struct {
	Message  string `json:"message"`
	Sender   uint   `json:"sender"`
	Receiver uint   `json:"receiver"`
}

func (m *PrivateMessage) UnmarshalJSON(b []byte) error {
	var data map[string]interface{}

	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	message := data["message"]
	sender := data["sender"]
	receiver := data["receiver"]

	if message == nil || message.(string) == "" {
		return errors.New("No message")
	}
	if sender == nil || sender.(float64) == 0 {
		return errors.New("No sender")
	}
	if receiver == nil || receiver.(float64) == 0 {
		return errors.New("No receiver")
	}

	m.Message = message.(string)
	m.Sender = uint(sender.(float64))
	m.Receiver = uint(receiver.(float64))

	return nil
}
