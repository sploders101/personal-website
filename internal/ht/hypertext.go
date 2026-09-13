package ht

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"text/template"

	"gitlab.com/sploders101/personal-website/internal/env"
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

type BasePageConfig struct {
	Devmode    bool
	DailyTheme string
}

// var themes = []string{
// 	"oceanblue-decor",
// 	"darkred-decor",
// 	"pink-decor",
// 	"seagreen-decor",
// 	"brown-decor",
// 	"electricblue-decor",
// }

func getBasePageConfig() (BasePageConfig, error) {
	// now := time.Now()
	// daysSinceEpoch := now.Unix() / 86400
	// dailyTheme := themes[daysSinceEpoch%int64(len(themes))]
	return BasePageConfig{
		Devmode:    env.Devmode,
		DailyTheme: "",
	}, nil
}

type basicTemplate struct {
	name string
}
func BasicTemplate(templateName string) http.Handler {
	return &basicTemplate{templateName}
}
func (template *basicTemplate) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	baseConfig, err := getBasePageConfig()
	if err != nil {
		slog.Error("Error generating base page config", "config", baseConfig)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "text/html")
	if err := templates.ExecuteTemplate(resp, template.name, baseConfig); err != nil {
		slog.Error("Failed to render page", "template", template.name, "error", err.Error())
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
}
