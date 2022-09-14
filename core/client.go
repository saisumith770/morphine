package core

import (
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

type Conn struct {
	id     string
	hub    *Hub
	topics []string //all topics a client is subscribed to
	send   chan []byte
	socket *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, //default read size
	WriteBufferSize: 1024, //default write size
	CheckOrigin: func(r *http.Request) bool {
		//have to implement a security protocol into this
		return true
	},
	HandshakeTimeout: time.Duration(time.Minute), // max of 1 minute to make the connection

	//later add errors and compressions
}

func Generate_ClientWS(
	h *Hub,
	req *http.Request,
	resp http.ResponseWriter,
) {
	conn, err := upgrader.Upgrade(resp, req, resp.Header())
	if err != nil {
		log.Fatal("WEBSOCKET::UPGRADER: failed to updgrade the websocket connection")
	}

	//generate uuid for the connection
	id, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal("UUID: couldn't generate uuid using os/exec")
	}

	client := &Conn{
		id:     string(id),
		hub:    h,
		topics: make([]string, 10), //can subscribe to a max of 10 topics
		send:   make(chan []byte),
		socket: conn,
	}

	client.socket.Close()
}
