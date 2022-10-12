package core

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

/*
leverages pub/sub mechanism
run a goroutine and use a channel to broadcast messages
*/

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
	disconnect chan *Conn
	webhook    chan WebhookConnInfo
	webhooks   map[string][]string
}

func Generate_HubService() (h *Hub) {
	return &Hub{
		clients:    make(map[string]*Conn),
		rooms:      make(map[string][]*Conn), //rooms are collections of Conn subscribed to some topic
		broadcast:  make(chan Message),
		join:       make(chan ChannelConnInfo),
		leave:      make(chan ChannelConnInfo),
		disconnect: make(chan *Conn),
		webhook:    make(chan WebhookConnInfo),
		webhooks:   make(map[string][]string),
	}
}

func remove[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (h *Hub) Run() {
	for {
		select {
		case channel_conn_info := <-h.join:
			//systems level topics require permission to join
			if channel_conn_info.client.topic_arr_len < 10 {
				var alreadySubscribed bool = false
				for _, topic := range channel_conn_info.client.topics {
					if topic == channel_conn_info.topic {
						alreadySubscribed = true
						break
					}
				}

				if !alreadySubscribed {
					channel_conn_info.client.topics = append(channel_conn_info.client.topics, channel_conn_info.topic)
					h.rooms[channel_conn_info.topic] = append(h.rooms[channel_conn_info.topic], channel_conn_info.client)
					channel_conn_info.client.topic_arr_len++

					log.Printf("ROOM::ACCESS: id:%v has joined room:%v", channel_conn_info.client.id, channel_conn_info.topic)
				} else {
					log.Printf("ROOM::ACCESS: id:%v already joined room:%v", channel_conn_info.client.id, channel_conn_info.topic)
				}
			}
		case channel_conn_info := <-h.leave:
			//using goroutines to asynchronise the process
			go func() {
				for index, conn := range h.rooms[channel_conn_info.topic] {
					if conn.id == channel_conn_info.client.id {
						h.rooms[channel_conn_info.topic] = remove(h.rooms[channel_conn_info.topic], index)
						if len(h.rooms[channel_conn_info.topic]) == 0 {
							delete(h.rooms, channel_conn_info.topic)
						}
						break
					}
				}

				log.Printf("ROOM::ACCESS: removed id:%v from array for room:%v", channel_conn_info.client.id, channel_conn_info.topic)
			}()
			go func() {
				for index, topic := range channel_conn_info.client.topics {
					if topic == channel_conn_info.topic {
						channel_conn_info.client.topics = remove(channel_conn_info.client.topics, index)
						channel_conn_info.client.topic_arr_len--
						break
					}
				}

				log.Printf("ROOM::ACCESS: removed room:%v from topics array for id:%v", channel_conn_info.topic, channel_conn_info.client.id)
			}()
		case payload := <-h.broadcast:
			var authorized bool = false
			for _, conn := range h.rooms[payload.topic] {
				if conn.id == payload.conn_id {
					authorized = true
					break
				}
			}

			if authorized {
				go func() {
					for _, conn := range h.rooms[payload.topic] {
						conn.writeToWs_readFromHub(payload, "morphine.message")
					}
				}()
				go func() {
					for _, webhook := range h.webhooks[payload.topic] {
						byteBody, err := json.Marshal(map[string]string{
							"topic":   payload.topic,
							"message": string(payload.message),
						})
						postBody := bytes.NewBuffer(byteBody)
						if err != nil {
							log.Printf("ROOM::WEBHOOKS: failed to send post request to webhook:%v", webhook)
						} else {
							resp, err := http.Post(webhook, "application/json", postBody)
							if err != nil {
								log.Printf("ROOM::WEBHOOKS: failed to send post request to webhook:%v err:%v", webhook, err)
							} else {
								log.Printf("ROOM::WEBHOOKS: response:%v", resp)
							}
						}
					}
				}()
			} else {
				log.Printf("ROOM::ACCESS: id:%v failed to broadcast message to room:%v", payload.conn_id, payload.topic)
			}
		case dm := <-h.dm:
			h.clients[dm.topic].writeToWs_readFromHub(dm, "morphine.direct_message")
		case webhook := <-h.webhook:
			h.webhooks[webhook.Topic] = append(h.webhooks[webhook.Topic], webhook.Url)
			log.Printf("WEBHOOK::CREATE: successfully subscribed url:%v to room:%v", webhook.Url, webhook.Topic)
		case conn := <-h.disconnect:
			for topic, room := range h.rooms {
				for index, connection := range room {
					if connection.id == conn.id {
						h.rooms[topic] = remove(room, index)
						if len(h.rooms[topic]) == 0 {
							delete(h.rooms, topic)
						}
						break
					}
				}
			}
			delete(h.clients, conn.name)
			log.Printf("CONN::STATE: disconnected %v from all rooms", conn.id)
		}
	}
}
