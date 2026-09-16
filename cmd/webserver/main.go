package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"

	// "github.com/yuin/goldmark"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sploders101/personal-website/cmd/webserver/ht"
	"github.com/sploders101/personal-website/cmd/webserver/userdata"
	"github.com/sploders101/personal-website/internal/config"
	"github.com/sploders101/personal-website/internal/dbapi"
	"github.com/sploders101/personal-website/internal/env"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load(configPath())
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("Loaded configuration")

	db, err := dbapi.NewDb(ctx, "pgx", cfg.Database.URL)
	if err != nil {
		slog.Error("Error opening connection to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	address := "[::]:8080"

	router := http.NewServeMux()
	router.Handle("GET /", http.FileServerFS(ht.StaticAssets))
	router.Handle("GET /{$}", userdata.UserMiddleware(db, ht.ServeHome(cfg)))
	router.Handle("GET /login/", userdata.UserMiddleware(db, ht.ServeLogin(cfg)))
	// router.Handle("POST /login/", http.HandlerFunc())
	if err := registerOidcHandlers(ctx, cfg, db, router); err != nil {
		slog.Error("Error registering oidc handlers", "error", err)
		os.Exit(1)
	}
	router.Handle("GET /decorations/{svgfile}", ht.DecorationServer{})

	// Enable hot reloading
	if env.Devmode {
		router.Handle(
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

	// Start server
	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("Error listening on specified address", "address", address, "error", err)
		os.Exit(1)
	}
	slog.Info("Listening for connections", "address", address)
	if err := http.Serve(listener, router); err != nil {
		slog.Error("Error serving connections", "address", address, "error", err)
		os.Exit(1)
	}
}

// configPath returns the path of the config file, overridable via CONFIG_FILE.
func configPath() string {
	if p := os.Getenv("CONFIG_FILE"); p != "" {
		return p
	}
	return "config.json"
}
