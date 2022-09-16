package core

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Conn struct {
	id     string
	hub    *Hub
	topics []string //all topics a client is subscribed to
	socket *websocket.Conn
}

type SocketMessagePayload struct {
	Topic       string `json:"topic"`
	Event       string `json:"event"`
	JsonMessage string `json:"json_message"`
}

func (c *Conn) readWsPayload_transferToHub() {
	defer func() {
		c.hub.disconnect <- c
		c.socket.Close()
	}()

	for {
		var payload SocketMessagePayload
		c.socket.ReadJSON(payload)

		switch payload.Event {
		case "morphine.join":
			c.hub.join <- ChannelConnInfo{
				topic:  payload.Topic,
				client: c,
			}
		case "morphine.leave":
			c.hub.leave <- ChannelConnInfo{
				topic:  payload.Topic,
				client: c,
			}
		case "morphine.message":
			c.hub.broadcast <- Message{
				topic:   payload.Topic,
				message: []byte(payload.JsonMessage),
			}
		default:
			c.writeToWs_readFromHub(Message{
				topic:   "system",
				message: []byte("custom events not supported yet. Use base events: 'morphine.join','morphine.leave','morphine.message'"),
			}, "morphine.invalid_event")
		}
	}
}

func (c *Conn) writeToWs_readFromHub(msg Message, event string) {
	var socketMessage SocketMessagePayload = SocketMessagePayload{
		Topic:       msg.topic,
		Event:       event,
		JsonMessage: string(msg.message),
	}

	if err := c.socket.WriteJSON(socketMessage); err != nil {
		log.Println("SOCKET::WRITE: error while writing a message to the websocket")
		return
	}
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
	socket, err := upgrader.Upgrade(resp, req, nil) //no automatic setting of response headers
	if err != nil {
		log.Fatal("WEBSOCKET::UPGRADER: failed to updgrade the websocket connection")
	}

	//generate uuid for the connection
	id := uuid.New()
	if err != nil {
		log.Fatal("UUID: couldn't generate uuid using os/exec")
	}

	conn := &Conn{
		id:     id.String(),
		hub:    h,
		topics: make([]string, 10), //can subscribe to a max of 10 topics
		socket: socket,
	}

	conn.readWsPayload_transferToHub()
}
