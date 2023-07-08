package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"context"

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
	members []map[string]interface{}
	// OpenAI API Token
	openaitoken string
}

func newRoom(logging bool, openaitoken string) *room {
	tracer := trace.New()
	if !logging {
		tracer = trace.Off()
	}
	return &room{
		forward:     make(chan *message),
		join:        make(chan *client),
		leave:       make(chan *client),
		clients:     make(map[*client]bool),
		tracer:      tracer,
		openaitoken: openaitoken,
	}
}

func (r *room) run() {
	// botのメッセージを保持するスライス
	var messages []*ChatMessage

	// OpenAIのクライアントを作成
	cli, err := NewClient(r.openaitoken)
	if err != nil {
		panic(err)
	}

	// 人格呼び出し
	personality := GetPersonality("zuri")
	ctx := context.Background()

	for {
		select {
		case client := <-r.join:
			// join this room
			r.clients[client] = true
			// keep state
			r.number++
			name := client.userData["name"].(string)
			remoteAddr := client.userData["remote_addr"].(string)
			r.members = append(
				r.members,
				objx.New(map[string]interface{}{
					"name":        name,
					"remote_addr": remoteAddr,
				}),
			)

			message := fmt.Sprintf("Joined a new member, %s (%s) !", name, remoteAddr)
			logContent := r.tracer.LogContent("join", name, remoteAddr, "-")
			r.tracer.TraceInfo(logContent)
			// send system message
			msg := makeSystemMessage(message)
			sendMessageAllClients(r, msg)
		case client := <-r.leave:
			// leave from room
			delete(r.clients, client)
			close(client.send)
			// keep state
			r.number--
			name := client.userData["name"].(string)
			remoteAddr := client.userData["remote_addr"].(string)

			r.members = remove(
				r.members,
				objx.New(map[string]interface{}{
					"name":        name,
					"remote_addr": remoteAddr,
				}),
			)

			message := fmt.Sprintf("%s (%s) left. Good bye.", name, remoteAddr)
			logContent := r.tracer.LogContent("leave", name, remoteAddr, "-")
			r.tracer.TraceInfo(logContent)
			msg := makeSystemMessage(message)
			sendMessageAllClients(r, msg)
		case msg := <-r.forward:
			if msg.Message != "pingpongpangpong" {
				logContent := r.tracer.LogContent("forward", msg.Name, msg.RemoteAddr, msg.Message)
				r.tracer.TraceInfo(logContent)
				sendMessageAllClients(r, msg)

				// botに反応させる
				// ユーザーの発言を受け取る
				userMessage := NewChatMessage(RoleUser, msg.Name, "")
				userMessage.Text = msg.Message
				messages = append(messages, userMessage)

				// OpenAIのAPIを叩いて、AIの発言を作成
				message, err := cli.Completion(ctx, msg.Name, personality, messages)
				if err != nil {
					fmt.Println("error:", err.Error())
					continue
				}
				fmt.Printf("[%s]\n", message)
				messages = append(messages, message)

				// bot のメッセージを送信
				msg.Name = "ずり"
				msg.Message = message.Text
				sendMessageAllClients(r, msg)
			}
		}
	}
}

func remove(maps []map[string]interface{}, search map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, v := range maps {
		if v["name"] == search["name"] && v["remote_addr"] == search["remote_addr"] {
			// no operation
		} else {
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
		default:
			// fail to sending message
			delete(r.clients, client)
			close(client.send)
			logContent := r.tracer.LogContent("system", msg.Name, msg.RemoteAddr, msg.Message)
			r.tracer.TraceError(logContent, errors.New("Failed to send a message. Cleanup client..."))
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
		logContent := r.tracer.LogContent("system", "-", req.Header.Get("X-Forwarded-For"), "ServeHTTP")
		r.tracer.TraceError(logContent, err)
		return
	}

	cookie, err := req.Cookie("yukizuri")
	if err != nil {
		logContent := r.tracer.LogContent("system", "-", req.Header.Get("X-Forwarded-For"), "Failed to get cookie data")
		r.tracer.TraceError(logContent, err)
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
