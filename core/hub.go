package core

/*
leverages pub/sub mechanism
run a goroutine and use a channel to broadcast messages
*/

type ChannelConnInfo struct {
	topic  string
	client *Conn
}

type Message struct {
	topic   string
	message []byte
}

type Hub struct {
	rooms     map[string][]*Conn
	broadcast chan Message
	join      chan ChannelConnInfo
	leave     chan ChannelConnInfo
}

func Generate_HubService() (h *Hub) {
	return &Hub{
		rooms:     make(map[string][]*Conn), //rooms are collections of Conn subscribed to some topic
		broadcast: make(chan Message),
		join:      make(chan ChannelConnInfo),
		leave:     make(chan ChannelConnInfo),
	}
}

func remove(s []*Conn, i int) []*Conn {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (h *Hub) Run() {
	for {
		select {
		case channel_conn_info := <-h.join:
			h.rooms[channel_conn_info.topic] = append(h.rooms[channel_conn_info.topic], channel_conn_info.client)
		case channel_conn_info := <-h.leave:
			for index, conn := range h.rooms[channel_conn_info.topic] {
				if conn.id == channel_conn_info.client.id {
					h.rooms[channel_conn_info.topic] = remove(h.rooms[channel_conn_info.topic], index)
					break
				}
			}
		case payload := <-h.broadcast:
			for _, conn := range h.rooms[payload.topic] {
				conn.send <- payload.message
			}
		}
	}
}
