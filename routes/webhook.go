package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"com.mixin.morphine/core"
)

func Webhook(hub *core.Hub) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		switch access_token {
		case os.Getenv("ADMIN_ACCESS"):
		case os.Getenv("SERVER_ACCESS"):
		default:
			res.WriteHeader(401)
			return
		}

		var webhook_payload core.WebhookConnInfo
		json.NewDecoder(req.Body).Decode(&webhook_payload)

		core.Subscribe_Webhook(hub, webhook_payload.Topic, webhook_payload.Url)
	}
}
