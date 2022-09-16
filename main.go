package main

import (
	"log"
	"net/http"
	"strings"

	"com.mixin.morphine/core"
	"github.com/gorilla/mux"
)

func main() {
	hub := core.Generate_HubService()
	router := mux.NewRouter()

	go hub.Run()
	router.HandleFunc("/connect", func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		log.Print(access_token)
		//check whether the proper access token is provided
		//access token must be stored with a salt
		core.Generate_ClientWS(hub, req, res)
	})

	log.Fatal("RUNNING::SERVER: ", http.ListenAndServe(":4001", router))
}
