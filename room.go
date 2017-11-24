package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/a-know/yukizuri/trace"
	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
)

type room struct {
	// channel for save message sending other
	forward chan *message
	// channel for client intend to join room. use for add client to `clients`
	join chan *client
	// channel for client intend to leave room. use for remove client from `clients`
	leave chan *client
	// save all joined clients
	clients map[*client]bool
	// logging
	tracer trace.Tracer
	// get avatar info
	avatar Avatar
	// joined members number
	number int
}

func newRoom(logging bool) *room {
	tracer := trace.New(os.Stdout)
	if !logging {
		tracer = trace.Off()
	}
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  tracer,
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// join this room
			r.clients[client] = true
			r.number++
			r.tracer.Trace(fmt.Sprintf("Joined a new client. Joined members count: %d", r.number))
		case client := <-r.leave:
			// leave from room
			delete(r.clients, client)
			close(client.send)
			r.number--
			r.tracer.Trace(fmt.Sprintf("Leave a client. Joined members count: %d", r.number))
		case msg := <-r.forward:
			r.tracer.Trace("Receive a message: ", msg.Message)
			// send message to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					// send a message
					r.tracer.Trace(" -- Send a message: ", msg.Message)
				default:
					// fail to sending message
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- Failed to send a message. Cleanup client...")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// for using websocket. Upgrade HTTP connection by websocket.Upgrader
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get cookie data:", err)
		return
	}
	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
