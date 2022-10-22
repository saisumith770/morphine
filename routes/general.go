package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"com.mixin.morphine/core"
	"github.com/go-redis/redis/v9"
)

var ctx = context.Background()

func ReadConnections(hub *core.Hub, rdb *redis.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		switch access_token {
		case os.Getenv("ADMIN_ACCESS"):
		case os.Getenv("SERVER_ACCESS"):
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
	}
}

func CreateWebsocket(hub *core.Hub, rdb *redis.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		profileDetails := core.WebsocketProfileDetails{
			Name:   req.URL.Query().Get("name"),
			Avatar: req.URL.Query().Get("avatar"),
		}
		var sessionid string
		var role core.Role
		switch access_token {
		case os.Getenv("ADMIN_ACCESS"):
			role = core.ADMIN
			if req.URL.Query().Get("sessionid") != "" {
				sessionid = req.URL.Query().Get("sessionid")
			} else {
				sessionid = "admin"
			}
		case os.Getenv("SERVER_ACCESS"):
			role = core.SERVER
			if req.URL.Query().Get("sessionid") != "" {
				sessionid = req.URL.Query().Get("sessionid")
			} else {
				sessionid = "server"
			}
		default:
			role = core.USER
			cookie, err := req.Cookie("sessionid")
			if err != nil {
				log.Printf("SERVER::COOKIE: could not access the sessionid cookie")
				res.WriteHeader(401)
				return
			}
			sessionid = cookie.Value
			details, err := rdb.Get(ctx, sessionid).Result()
			if err != nil {
				log.Printf("REDIS::KEY: could not get the session id %v from redis", sessionid)
				res.WriteHeader(401)
				return
			}
			json.Unmarshal([]byte(details), &profileDetails)
		}
		core.Generate_ClientWS(hub, req, res, profileDetails, role, sessionid)
	}
}
