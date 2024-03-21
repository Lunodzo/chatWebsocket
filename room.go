package main

import (
	"log"
	"github.com/gorilla/websocket"
	"net/http"
)

type room struct {
	//Hold all clients in this room
	clients map[*client]bool

	//Channel for clients to join the room
	join chan *client

	//Channel for clients to leave the room
	leave chan *client

	//Incoming messages that should be forwarded to other clients
	forward chan []byte
}

//Create a new room
func newRoom() *room {
	return &room{
		clients: make(map[*client]bool),
		join: make(chan *client),
		leave: make(chan *client),
		forward: make(chan []byte),
	}
}

//Run the room
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			//Joining
			r.clients[client] = true
		case client := <-r.leave:
			//Leaving
			delete(r.clients, client)
			close(client.receive)
		case msg := <-r.forward:
			//Forward message to all clients
			for client := range r.clients {
				select {
				case client.receive <- msg:
					//Send the message
				default:
					//Failed to send
					delete(r.clients, client)
					close(client.receive)
				}
			}
		}
	}
}

const(
	socketBufferSize = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	client := &client{
		socket: socket,
		receive: make(chan []byte, messageBufferSize),
		room: r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}