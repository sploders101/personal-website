package ht

import (
	"context"
	"embed"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"text/template"

	"github.com/sploders101/personal-website/cmd/webserver/userdata"
	"github.com/sploders101/personal-website/internal/config"
	queries "github.com/sploders101/personal-website/internal/dbapi/gen"
	"github.com/sploders101/personal-website/internal/env"
)

//go:embed templates assets
var assets embed.FS

var templates *template.Template

func init() {
	templates = template.Must(template.ParseFS(
		assets,
		"templates/html/components/*.html",
	))
	template.Must(templates.ParseFS(
		assets,
		"templates/html/pages/*.html",
	))
}

type baseTemplateVars struct {
	Devmode    bool
	DailyTheme string
	Config     config.ServerConfig
	User       queries.User
}

// var themes = []string{
// 	"oceanblue-decor",
// 	"darkred-decor",
// 	"pink-decor",
// 	"seagreen-decor",
// 	"brown-decor",
// 	"electricblue-decor",
// }

func getBasePageConfig(ctx context.Context, cfg config.ServerConfig) (baseTemplateVars, error) {
	// now := time.Now()
	// daysSinceEpoch := now.Unix() / 86400
	// dailyTheme := themes[daysSinceEpoch%int64(len(themes))]
	return baseTemplateVars{
		Devmode:    env.Devmode,
		DailyTheme: "",
		Config:     cfg,
		User:       userdata.GetUserData(ctx),
	}, nil
}

func BaseTemplate(cfg config.ServerConfig, templateName string) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		baseCfg, err := getBasePageConfig(req.Context(), cfg)
		if err != nil {
			slog.Error("Error generating base page config", "config", baseCfg)
			http.Error(resp, "Internal server error", http.StatusInternalServerError)
			return
		}

		resp.Header().Set("Content-Type", "text/html")
		if err := templates.ExecuteTemplate(resp, templateName, baseCfg); err != nil {
			slog.Error("Failed to render page", "template", templateName, "error", err)
			http.Error(resp, "Internal server error", http.StatusInternalServerError)
			return
		}
	})
}

func ServeAssets(cfg config.ServerConfig) http.Handler {
	handler404 := Serve404(cfg)
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		file, err := assets.Open(path.Join("assets", req.URL.Path))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// Serve 404
				handler404.ServeHTTP(resp, req)
				return
			}
			// Log err, serve 500
			slog.Error("Failed to load file", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		stat, err := file.Stat()
		if err != nil {
			slog.Error("Failed to stat file", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Serve file
		http.ServeContent(resp, req, path.Base(req.URL.Path), stat.ModTime(), file.(io.ReadSeeker))
	})
}

func ServeHome(cfg config.ServerConfig) http.Handler {
	return BaseTemplate(cfg, "home.html")
}

func ServeLogin(cfg config.ServerConfig) http.Handler {
	return BaseTemplate(cfg, "login.html")
}

func Serve404(cfg config.ServerConfig) http.Handler {
	return BaseTemplate(cfg, "404.html")
}
