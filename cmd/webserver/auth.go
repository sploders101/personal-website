package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/sploders101/personal-website/cmd/webserver/dbapi"
	"github.com/sploders101/personal-website/cmd/webserver/userdata"
	"github.com/sploders101/personal-website/internal/env"
)

func serveLogout(db dbapi.Db) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		tx, err := db.Begin(req.Context())
		if err != nil {
			slog.Error("Failed to open database transaction", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()
		userSession := userdata.GetSessionInfo(req.Context())
		if err := db.Query().DeleteUserSession(req.Context(), userSession.TokenHash); err != nil {
			slog.Error("Failed to delete user session", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(); err != nil {
			slog.Error("Failed to commit transaction", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Delete cookie by updating it with a `0` expiration time
		http.SetCookie(resp, &http.Cookie{
			Name:     "shaunkeyscom-session",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   !env.Devmode,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		})

		http.Redirect(resp, req, "/", http.StatusFound)
	})
}
