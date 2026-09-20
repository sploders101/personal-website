package helpers

import (
	"net/http"

	"github.com/sploders101/personal-website/cmd/webserver/userdata"
)

func RequireLogin(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		userData := userdata.GetUserData(req.Context())
		if userData.ID == 0 {
			http.Redirect(resp, req, "/login/", http.StatusFound)
			return
		}
		handler.ServeHTTP(resp, req)
	})
}
