package main

import (
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
		core.Generate_ClientWS(hub, req, res)
	})

	log.Fatal("RUNNING::SERVER: ", http.ListenAndServe(":4001", router))
}
