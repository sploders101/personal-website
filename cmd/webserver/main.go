package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sort"

	// "github.com/yuin/goldmark"

	"github.com/sploders101/personal-website/internal/config"
	"github.com/sploders101/personal-website/internal/env"
	"github.com/sploders101/personal-website/internal/ht"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load(configPath())
	if err != nil {
		slog.Error("Error loading configuration", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Loaded configuration")

	address := "[::]:8080"

	router := http.NewServeMux()
	router.Handle("GET /", http.FileServerFS(ht.StaticAssets))
	router.Handle("GET /{$}", ht.ServeHome(cfg))
	router.Handle("GET /login/", ht.ServeLogin(cfg))
	// router.Handle("POST /login/", http.HandlerFunc())
	registerOidcHandlers(ctx, cfg, router)
	router.Handle("GET /decorations/{svgfile}", ht.DecorationServer{})

	// Enable hot reloading
	if env.Devmode {
		router.Handle("GET /_reload", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.Header().Set("Content-Type", "text/event-stream")
			resp.Header().Set("Cache-Control", "no-cache")
			resp.Header().Set("Connection", "keep-alive")
			flusher, ok := resp.(http.Flusher)
			if !ok {
				http.Error(resp, "Streaming unsupported", http.StatusInternalServerError)
				return
			}
			resp.Write([]byte("data: online\n\n"))
			flusher.Flush()
			<-req.Context().Done()
		}))
	}

	// Start server
	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("Error listening on specified address", "address", address, "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Listening for connections", "address", address)
	if err := http.Serve(listener, router); err != nil {
		slog.Error("Error serving connections", "address", address, "error", err.Error())
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

// backendNames returns the configured storage backend names in sorted order.
func backendNames(storage map[string]config.StorageBackend) []string {
	names := make([]string, 0, len(storage))
	for name := range storage {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
