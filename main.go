package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	// "github.com/yuin/goldmark"
	"html/template"
)

var templates *template.Template

func init() {
	templates = template.Must(template.ParseGlob(
		"./templates/html/components/*.html",
	))
	template.Must(templates.ParseGlob(
		"./templates/html/pages/*.html",
	))
}

func main() {
	address := "[::]:8080"
	assetsPath := "./assets"
	assetFolder, err := os.OpenRoot(assetsPath)
	if err != nil {
		slog.Error("Error opening root.", "root", assetsPath, "error", err.Error())
		os.Exit(1)
	}

	router := http.NewServeMux()
	assetServer := http.FileServerFS(assetFolder.FS())
	router.Handle("GET /{$}", http.HandlerFunc(serveHomepage))
	router.Handle("GET /", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		assetServer.ServeHTTP(resp, req)
	}))
	router.Handle("GET /decorations/{svgfile}", http.HandlerFunc(serveDecoration))

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

type BasePageConfig struct {
	DailyTheme string
}

var themes = []string{
	"oceanblue-decor",
	"darkred-decor",
	"pink-decor",
	"seagreen-decor",
	"brown-decor",
	"electricblue-decor",
}

func getBasePageConfig() (BasePageConfig, error) {
	now := time.Now()
	daysSinceEpoch := now.Unix() / 86400
	dailyTheme := themes[daysSinceEpoch%int64(len(themes))]
	return BasePageConfig{
		DailyTheme: dailyTheme,
	}, nil
}

func serveHomepage(resp http.ResponseWriter, req *http.Request) {
	baseConfig, err := getBasePageConfig()
	if err != nil {
		slog.Error("Error generating base page config", baseConfig)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "text/html")
	if err := templates.ExecuteTemplate(resp, "home.html", baseConfig); err != nil {
		slog.Error("Failed to render home page", "error", err.Error())
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
}
