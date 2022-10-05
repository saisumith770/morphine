package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"com.mixin.morphine/core"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("SYSTEM::ENVIRONMENT: failed to load environment variables from godotenv")
	}
	hub := core.Generate_HubService()
	router := mux.NewRouter()

	go hub.Run()

	router.HandleFunc("/analytics", func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		if access_token == os.Getenv("ADMIN_ACCESS") {
			//provide server analytics here
		} else {
			res.WriteHeader(401)
		}
	}).Methods("GET")

	router.HandleFunc("/webhook", func(res http.ResponseWriter, req *http.Request) {
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
	})

	router.HandleFunc("/connections", func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		switch access_token {
		case os.Getenv("ADMIN_ACCESS"):
		case os.Getenv("SERVER_ACCESS"):
		case os.Getenv("USER_ACCESS"):
		default:
			res.WriteHeader(401)
			return
		}

		core.Room_Presence(hub, req.URL.Query().Get("room"))
	})

	router.HandleFunc("/connect", func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		switch access_token {
		case os.Getenv("ADMIN_ACCESS"):
		case os.Getenv("SERVER_ACCESS"):
		case os.Getenv("USER_ACCESS"):
		default:
			res.WriteHeader(401)
			return
		}

		profileDetails := core.WebsocketProfileDetails{
			Name:   req.URL.Query().Get("name"),
			Avatar: req.URL.Query().Get("avatar"),
		}
		// json.NewDecoder(req.Body).Decode(&profileDetails)

		core.Generate_ClientWS(hub, req, res, profileDetails)
	})

	log.Fatal("RUNNING::SERVER: ", http.ListenAndServe(":4001", router))
}
