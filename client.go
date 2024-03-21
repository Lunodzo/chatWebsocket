package main 

import (
	"github.com/gorilla/websocket"
)

type client struct {
	//web socket for this client
	socket *websocket.Conn

	//channel on which messages are received
	receive chan []byte

	//room in which the client is chatting
	room *room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}