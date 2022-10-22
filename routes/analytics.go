package routes

import (
	"net/http"
	"os"
	"strings"
)

func ServerAnalytics() func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		access_token := strings.TrimPrefix(req.Header.Get("Access-Token"), "Bearer ")
		if access_token == os.Getenv("ADMIN_ACCESS") {
			//provide server analytics here
		} else {
			res.WriteHeader(401)
		}
	}
}
