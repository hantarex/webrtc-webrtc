package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"webrtc-webrtc/gstreamer"
)

type WebSocket struct {
	*websocket.Conn
	gstreamer.GStreamer
	Errs chan string
}

func (self *WebSocket) ReadMessages() {
	var msg gstreamer.Message
	_, message, err := self.ReadMessage()

	if err != nil {
		log.Println("read:", err)
		self.Conn.Close()
		return
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Сбой демаршалинга JON: %s\n", err)
		self.Conn.Close()
	}
	switch msg.Id {
	case "type":
		switch msg.Key {
		case "client":
			go self.readMessagesClient()
			break
		case "server":
			go self.readMessagesServer()
			break
		default:
			log.Println("Type of client not found")
			self.Conn.Close()
		}
		break
	default:
		log.Println(msg)
		log.Println("Error read type of client")
		self.Conn.Close()
	}
}

func (self *WebSocket) readMessagesClient() {
	for {
		var msg gstreamer.Message
		_, message, err := self.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			self.Conn.Close()
			return
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Сбой демаршалинга JON: %s\n", err)
			self.Conn.Close()
		}
		switch msg.Id {
		case "client_answer":
			if err := self.GStreamer.On_answer_received(msg, self.GStreamer.Webrtc); err != nil {
				log.Println(err.Error())
			}
			break
		case "onIceCandidateClient":
			self.GStreamer.IceCandidateReceived(msg, self.GStreamer.Webrtc)
			break
		default:
			log.Println(msg)
			log.Println("Error readMessages")
		}
	}
}

func (self *WebSocket) readMessagesServer() {
	self.InitGstServer()
	for {
		var msg gstreamer.Message
		_, message, err := self.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			self.Conn.Close()
			return
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Сбой демаршалинга JON: %s\n", err)
			self.Conn.Close()
		}
		switch msg.Id {
		case "client_answer":
			if err := self.GStreamer.On_answer_received(msg, self.GStreamer.Webrtc); err != nil {
				log.Println(err.Error())
			}
			break
		case "onIceCandidateClient":
			self.GStreamer.IceCandidateReceived(msg, self.GStreamer.Webrtc)
			break
		default:
			log.Println(msg)
			log.Println("Error readMessages")
		}
	}
}

func (self *WebSocket) Ping() {
	for {
		select {
		case <-self.Errs:
			return
		case <-time.After(time.Second):
		}
		if err := self.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			return
		}
	}
}
