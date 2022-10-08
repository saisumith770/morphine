package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"com.mixin.morphine/core"
	"github.com/go-redis/redis/v9"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var ctx = context.Background()

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("SYSTEM::ENVIRONMENT: failed to load environment variables from godotenv")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

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
			cookie, err := req.Cookie("sessionid")
			if err != nil {
				log.Printf("SERVER::COOKIE: could not access the sessionid cookie")
				res.WriteHeader(401)
				return
			}
			_, cookieErr := rdb.Get(ctx, cookie.Value).Result()
			if cookieErr != nil {
				log.Printf("REDIS::KEY: could not get the session id %v from redis", cookie.Value)
				res.WriteHeader(401)
				return
			}
		}

		core.Room_Presence(hub, req.URL.Query().Get("room"))
	})

	router.HandleFunc("/connect", func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		profileDetails := core.WebsocketProfileDetails{
			Name:   req.URL.Query().Get("name"),
			Avatar: req.URL.Query().Get("avatar"),
		}
		switch access_token {
		case os.Getenv("ADMIN_ACCESS"):
		case os.Getenv("SERVER_ACCESS"):
		case os.Getenv("USER_ACCESS"):
		default:
			cookie, err := req.Cookie("sessionid")
			if err != nil {
				log.Printf("SERVER::COOKIE: could not access the sessionid cookie")
				res.WriteHeader(401)
				return
			}
			details, err := rdb.Get(ctx, cookie.Value).Result()
			if err != nil {
				log.Printf("REDIS::KEY: could not get the session id %v from redis", cookie.Value)
				res.WriteHeader(401)
				return
			}
			json.Unmarshal([]byte(details), &profileDetails)
		}
		core.Generate_ClientWS(hub, req, res, profileDetails)
	})

	log.Fatal("RUNNING::SERVER: ", http.ListenAndServe(":4001", router))
}
