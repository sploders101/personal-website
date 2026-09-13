package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"

	// "github.com/yuin/goldmark"

	"github.com/sploders101/personal-website/internal/env"
	"github.com/sploders101/personal-website/internal/ht"
	"github.com/jackc/pgx/v5"
)

func main() {
	address := "[::]:8080"

	router := http.NewServeMux()
	router.Handle("GET /", http.FileServerFS(ht.StaticAssets))
	router.Handle("GET /{$}", ht.BasicTemplate("home.html"))
	router.Handle("GET /login/", ht.BasicTemplate("login.html"))
	// router.Handle("POST /login/", http.HandlerFunc())
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

