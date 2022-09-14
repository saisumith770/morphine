package main

import (
	"net/http"

	"com.mixin.morphine/core"
	"github.com/gorilla/mux"
)

func main() {
	hub := core.Generate_HubService()
	router := mux.NewRouter()

	go hub.Run()
	router.HandleFunc("/connect", func(res http.ResponseWriter, req *http.Request) {
		core.Generate_ClientWS(hub, req, res)
	})

	http.ListenAndServe(":4001", router)
}
