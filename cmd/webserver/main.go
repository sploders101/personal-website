package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// "github.com/yuin/goldmark"

	"github.com/sploders101/personal-website/cmd/webserver/config"
	"github.com/sploders101/personal-website/cmd/webserver/dbapi"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load(configPath())
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	slog.Info("Loaded configuration")

	db, err := dbapi.NewDb(ctx, cfg.Database.URL)
	if err != nil {
		slog.Error("Error opening connection to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	address := "[::]:8080"

	router := http.NewServeMux()
	router.Handle("/", makeWebRouter(ctx, cfg, db))

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
