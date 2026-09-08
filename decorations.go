package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"regexp"
)

var svgtmpl *template.Template
func init() {
	svgtmpl = template.Must(template.ParseGlob("templates/decorations/*.svg"))
}

type DecorationConfig struct {
	Color string
}

func serveDecoration(resp http.ResponseWriter, req *http.Request) {
	svgfile := req.PathValue("svgfile")
	query := req.URL.Query()

	color := "#" + query.Get("color")
	if !safeHexColor(color) {
		resp.Header().Add("Content-Type", "text/plain")
		resp.WriteHeader(400)
		resp.Write([]byte("Invalid color"))
		return
	}

	resp.Header().Set("Content-Type", "image/svg+xml")
	resp.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'inline';")
	if err := svgtmpl.ExecuteTemplate(resp, svgfile, DecorationConfig{
		Color: color,
	}); err != nil {
		slog.Error("Failed to render SVG template", "file", svgfile, "error", err.Error())
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}
}

var hexRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func safeHexColor(s string) bool {
	if len(s) != 7 {
		return false
	}
	if s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
