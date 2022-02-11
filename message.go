package main

type Message struct {
	Type     string
	Message  string
	Channel  uint
	Receiver uint
	Sender   uint
	Socket   *Socket
}

func Broadcast(msg string, channel uint, sender uint) Message {
	return Message{
		Type:    "broadcast",
		Message: msg,
		Channel: channel,
		Sender:  sender,
	}
}

func SendPrivate(msg string, receiver uint, sender uint) Message {
	return Message{
		Type:     "priv_msg",
		Message:  msg,
		Sender:   sender,
		Receiver: receiver,
	}
}

func JoinChannel(channel uint) Message {
	return Message{
		Type:    "join_channel",
		Channel: channel,
	}
}
