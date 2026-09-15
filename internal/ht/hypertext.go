package ht

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/sploders101/personal-website/internal/config"
	"github.com/sploders101/personal-website/internal/env"
)

//go:embed templates assets
var assets embed.FS

var StaticAssets fs.FS

func init() {
	assetFolder, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	StaticAssets = assetFolder
}

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
}

// var themes = []string{
// 	"oceanblue-decor",
// 	"darkred-decor",
// 	"pink-decor",
// 	"seagreen-decor",
// 	"brown-decor",
// 	"electricblue-decor",
// }

func getBasePageConfig(cfg config.ServerConfig) (baseTemplateVars, error) {
	// now := time.Now()
	// daysSinceEpoch := now.Unix() / 86400
	// dailyTheme := themes[daysSinceEpoch%int64(len(themes))]
	return baseTemplateVars{
		Devmode:    env.Devmode,
		DailyTheme: "",
		Config:     cfg,
	}, nil
}

func BaseTemplate(cfg config.ServerConfig, templateName string) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		baseCfg, err := getBasePageConfig(cfg)
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

func ServeHome(cfg config.ServerConfig) http.Handler {
	return BaseTemplate(cfg, "home.html")
}

func ServeLogin(cfg config.ServerConfig) http.Handler {
	return BaseTemplate(cfg, "login.html")
}
