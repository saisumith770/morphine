package core

import "log"

/*
leverages pub/sub mechanism
run a goroutine and use a channel to broadcast messages
*/

type ChannelConnInfo struct {
	topic  string
	client *Conn
}

type Message struct {
	conn_id string
	topic   string
	message []byte
}

type Hub struct {
	rooms      map[string][]*Conn
	broadcast  chan Message
	join       chan ChannelConnInfo
	leave      chan ChannelConnInfo
	disconnect chan *Conn
}

func Generate_HubService() (h *Hub) {
	return &Hub{
		rooms:      make(map[string][]*Conn), //rooms are collections of Conn subscribed to some topic
		broadcast:  make(chan Message),
		join:       make(chan ChannelConnInfo),
		leave:      make(chan ChannelConnInfo),
		disconnect: make(chan *Conn),
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
				for _, conn := range h.rooms[payload.topic] {
					conn.writeToWs_readFromHub(payload, "morphine.message")
				}
			} else {
				log.Printf("ROOM::ACCESS: id:%v failed to broadcast message to room:%v", payload.conn_id, payload.topic)
			}
		case conn := <-h.disconnect:
			for topic, room := range h.rooms {
				for index, connection := range room {
					if connection.id == conn.id {
						h.rooms[topic] = remove(room, index)
						break
					}
				}
			}
		}
	}
}
