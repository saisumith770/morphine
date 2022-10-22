package core

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (c *Conn) readWsPayload_transferToHub() {
	defer func() {
		c.hub.disconnect <- c
		c.socket.Close()
	}()

	readSocket := true
	for readSocket {
		var payload SocketMessagePayload
		c.socket.ReadJSON(&payload)

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
				conn_id: c.id,
				name:    c.name,
				avatar:  c.avatar,
				topic:   payload.Topic,
				message: []byte(payload.JsonMessage),
			}
		case "morphine.direct_message":
			if c.role == ADMIN || c.role == SERVER {
				c.hub.dm <- Message{
					conn_id: c.id,
					name:    c.name,
					avatar:  c.avatar,
					topic:   payload.Topic,
					message: []byte(payload.JsonMessage),
				}
			} else {
				log.Printf("SOCKET::EVENT: id:%v unauthorised to send direct message", c.id)
				c.writeToWs_readFromHub(Message{
					topic:   "system",
					message: []byte("connection not authorised to send direct message"),
				}, "morphine.unauthorised")
			}
		case "":
			log.Printf("CONN::STATE: id:%v closed due to empty event", c.id)
			readSocket = false //socket is most likely compromised
		default:
			log.Printf("SOCKET::EVENT: unknown event %v was received", payload.Event)
			c.writeToWs_readFromHub(Message{
				topic:   "system",
				message: []byte("custom events not supported yet. Use base events: 'morphine.join','morphine.leave','morphine.message'"),
			}, "morphine.invalid_event")
		}
	}
}

func (c *Conn) writeToWs_readFromHub(msg Message, event string) {
	var socketMessage SocketMessagePayload = SocketMessagePayload{
		Name:        msg.name,
		Avatar:      msg.avatar,
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
	profileDetails WebsocketProfileDetails,
	role Role,
	sessionid string,
) {
	socket, err := upgrader.Upgrade(resp, req, nil) //no automatic setting of response headers
	if err != nil {
		log.Printf("WEBSOCKET::UPGRADER: failed to updgrade the websocket connection")
	}

	//generate uuid for the connection
	id := uuid.New()
	if err != nil {
		log.Printf("UUID: couldn't generate uuid using google/uuid")
	}

	conn := &Conn{
		id:            id.String(),
		name:          profileDetails.Name,
		avatar:        profileDetails.Avatar,
		hub:           h,
		topics:        make([]string, 10), //can subscribe to a max of 10 topics
		topic_arr_len: 0,
		socket:        socket,
		role:          role,
	}

	h.clients[sessionid] = conn

	conn.readWsPayload_transferToHub()
}
