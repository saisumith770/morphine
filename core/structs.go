package core

import "github.com/gorilla/websocket"

//---client-structs---

type Role int64

const (
	ADMIN Role = iota
	SERVER
	USER
)

type Conn struct {
	id            string
	name          string
	avatar        string
	hub           *Hub
	topics        []string //all topics a client is subscribed to
	topic_arr_len int
	socket        *websocket.Conn
	role          Role
}

type SocketMessagePayload struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Topic       string `json:"topic"`
	Event       string `json:"event"`
	JsonMessage string `json:"json_message"`
}

type WebsocketProfileDetails struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

//---server-structs---

type ChannelConnInfo struct {
	topic  string
	client *Conn
}

type WebhookConnInfo struct {
	Topic string `json:"topic"`
	Url   string `json:"url"`
}

type Message struct {
	conn_id string
	name    string
	avatar  string
	topic   string
	message []byte
}

type Hub struct {
	clients    map[string]*Conn
	rooms      map[string][]*Conn
	broadcast  chan Message
	dm         chan Message
	join       chan ChannelConnInfo
	leave      chan ChannelConnInfo
	presence chan Message
	disconnect chan *Conn
	webhook    chan WebhookConnInfo
	webhooks   map[string][]string
}