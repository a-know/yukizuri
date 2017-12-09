package main

import (
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	socket *websocket.Conn
	// channel for sending message
	send     chan *message
	room     *room
	userData map[string]interface{}
}

func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.RemoteAddr = c.socket.RemoteAddr().String()
			msg.Name = c.userData["name"].(string)
			if avatarURL, ok := c.userData["avatar_url"]; ok {
				msg.AvatarURL = avatarURL.(string)
			}
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
