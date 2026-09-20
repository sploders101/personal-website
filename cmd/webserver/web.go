package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
	"github.com/sploders101/personal-website/cmd/webserver/config"
	"github.com/sploders101/personal-website/cmd/webserver/dbapi"
	"github.com/sploders101/personal-website/cmd/webserver/ht"
	"github.com/sploders101/personal-website/cmd/webserver/userdata"
	"github.com/sploders101/personal-website/internal/env"
)

func makeWebRouter(ctx context.Context, cfg config.ServerConfig, db dbapi.Db) http.Handler {
	webRouter := http.NewServeMux()
	webRouter.Handle("GET /", ht.ServeAssets(cfg, db))
	webRouter.Handle("GET /{$}", userdata.UserMiddleware(db, ht.ServeHome(cfg)))
	webRouter.Handle("GET /login/", userdata.UserMiddleware(db, ht.ServeLogin(cfg)))
	webRouter.Handle("POST /logout/", userdata.UserMiddleware(db, serveLogout(db)))
	webRouter.Handle("GET /profile/", userdata.UserMiddleware(db, ht.ServeProfile(cfg)))
	webRouter.Handle("GET /profile/edit/", userdata.UserMiddleware(db, ht.ServeProfileEdit(cfg)))
	webRouter.Handle("POST /profile/edit/", userdata.UserMiddleware(db, editUserProfile(db)))
	webRouter.Handle("POST /auth/local/firstfactor", serveLocalLogin(cfg, db))
	if err := registerOidcHandlers(ctx, cfg, db, webRouter); err != nil {
		slog.Error("Error registering oidc handlers", "error", err)
		os.Exit(1)
	}
	webRouter.Handle("GET /decorations/{svgfile}", ht.DecorationServer{})

	// Enable hot reloading
	if env.Devmode {
		webRouter.Handle(
			"GET /_reload",
			http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				resp.Header().Set("Content-Type", "text/event-stream")
				resp.Header().Set("Cache-Control", "no-cache")
				resp.Header().Set("Connection", "keep-alive")
				flusher, ok := resp.(http.Flusher)
				if !ok {
					http.Error(resp, "Streaming unsupported", http.StatusInternalServerError)
					return
				}
				if _, err := resp.Write([]byte("data: online\n\n")); err != nil {
					return
				}
				flusher.Flush()
				<-req.Context().Done()
			}),
		)
	}

	if env.Devmode {
		wrapped := csrf.Protect([]byte(cfg.Secrets.CsrfSecret), csrf.Secure(false))(webRouter)
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			wrapped.ServeHTTP(resp, csrf.PlaintextHTTPRequest(req))
		})
	} else {
		return csrf.Protect([]byte(cfg.Secrets.CsrfSecret))(webRouter)
	}
}
