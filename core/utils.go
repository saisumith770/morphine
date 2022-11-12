package core

func remove[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func Subscribe_Webhook(hub *Hub, topic string, url string) {
	hub.webhook <- WebhookConnInfo{
		Topic: topic,
		Url:   url,
	}
}

func Room_Presence(hub *Hub, topic string) []WebsocketProfileDetails {
	var result []WebsocketProfileDetails
	for _, conn := range hub.rooms[topic] {
		result = append(result, WebsocketProfileDetails{
			Name:   conn.name,
			Avatar: conn.avatar,
		})
	}
	return result
}
