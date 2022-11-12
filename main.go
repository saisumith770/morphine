package main

import (
	"log"
	"net/http"

	"com.mixin.morphine/core"
	"com.mixin.morphine/routes"
	"github.com/go-redis/redis/v9"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// work on implementations
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

	router.HandleFunc("/analytics", routes.ServerAnalytics()).Methods("GET")
	router.HandleFunc("/webhook", routes.Webhook(hub)).Methods("POST")
	router.HandleFunc("/connections", routes.ReadConnections(hub, rdb)).Methods("GET")
	router.HandleFunc("/connect", routes.CreateWebsocket(hub, rdb))

	log.Printf("RUNNING::SERVER: http://localhost:4001/")
	err := http.ListenAndServe(":4001", router)
	if err != nil {
		log.Fatal("RUNNING::SERVER: ", err)
	}
}
