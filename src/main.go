package main

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"webrtc-webrtc/gstreamer"
	websocket2 "webrtc-webrtc/websocket"
)

var useAddr, useRTMP string
var addrDockerWS = os.Getenv("WS_PORT")
var addrDockerRTMP = os.Getenv("RTMP_DST")
var addr = "8083"
var rtmp = "rtmp://127.0.0.1:1945/"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	ws := websocket2.WebSocket{Conn: c, Errs: make(chan string, 1), GStreamer: gstreamer.GStreamer{C: c}}
	go ws.ReadMessages()
	go ws.Ping()
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
}

func main() {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	if useAddr = addrDockerWS; addrDockerWS == "" {
		log.Printf("Not use env WS_PORT. Set default ws port: %s\n", addr)
		useAddr = addr
	}
	if useRTMP = addrDockerRTMP; addrDockerRTMP == "" {
		log.Printf("Not use env RTMP_DST. Set default addres: %s\n", rtmp)
		useRTMP = rtmp
	}
	log.Println("WS_PORT = " + useAddr + ". RTMP_DST = " + useRTMP)
	http.HandleFunc("/ws", ws)
	log.Printf("Server listen %s\n", ":"+useAddr)
	if err := http.ListenAndServe(":"+useAddr, nil); err != nil {
		log.Fatalln(err)
	}
}
