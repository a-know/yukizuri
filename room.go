package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
	// joined members number
	number int
	// joined members nickname slice
	members []string
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
			// keep state
			r.number++
			r.members = append(r.members, client.userData["name"].(string))

			message := fmt.Sprintf("Joined a new client. Joined members count: %d", r.number)
			r.tracer.Trace(message)
			// send system message
			msg := makeSystemMessage(message)
			sendMessageAllClients(r, msg)
		case client := <-r.leave:
			// leave from room
			delete(r.clients, client)
			close(client.send)
			// keep state
			r.number--
			r.members = remove(r.members, client.userData["name"].(string))

			message := fmt.Sprintf("Leave a client. Joined members count: %d", r.number)
			r.tracer.Trace(message)
			msg := makeSystemMessage(message)
			sendMessageAllClients(r, msg)
		case msg := <-r.forward:
			r.tracer.Trace("Receive a message: ", msg.Message)
			sendMessageAllClients(r, msg)
		}
	}
}

func remove(strings []string, search string) []string {
	result := []string{}
	for _, v := range strings {
		if v != search {
			result = append(result, v)
		}
	}
	return result
}

func makeSystemMessage(content string) *message {
	msg := &message{
		When:    time.Now(),
		Name:    "Yukizuri-sys",
		Message: content,
	}
	return msg
}

func sendMessageAllClients(r *room, msg *message) {
	msg.CurrentMembers = r.members
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

	cookie, err := req.Cookie("yukizuri")
	if err != nil {
		log.Fatal("Failed to get cookie data:", err)
		return
	}
	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(cookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
