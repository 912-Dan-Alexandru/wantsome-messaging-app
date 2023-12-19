package models

import (
	"time"

	"github.com/gorilla/websocket"
)

type Room struct {
	Name    string
	Members map[*websocket.Conn]bool
}

type Message struct {
	Message   string
	UserName  string
	Recipient string
	Room      string
	Type      string
	TimeStamp time.Time
}
