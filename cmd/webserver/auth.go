package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/sploders101/personal-website/cmd/webserver/config"
	"github.com/sploders101/personal-website/cmd/webserver/dbapi"
	queries "github.com/sploders101/personal-website/cmd/webserver/dbapi/gen"
	"github.com/sploders101/personal-website/cmd/webserver/userdata"
	"github.com/sploders101/personal-website/internal/env"
)

const (
	argon2Time    = 2
	argon2Memory  = 19 * 1024
	argon2Threads = 1
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

var argon2Params = &argon2id.Params{
	Memory:      argon2Memory,
	Iterations:  argon2Time,
	Parallelism: argon2Threads,
	SaltLength:  argon2SaltLen,
	KeyLength:   argon2KeyLen,
}

// dummyHash is verified against when a username does not exist, so that
// rejected logins take a similar amount of time whether or not the username
// exists. This mitigates user enumeration through response timing.
var dummyHash = mustHashPassword("shaunkeyscom-local-login-dummy")

func mustHashPassword(password string) string {
	encoded, err := hashPassword(password)
	if err != nil {
		panic(fmt.Sprintf("Failed to hash dummy password: %v", err))
	}
	return encoded
}

func serveLocalLogin(cfg config.ServerConfig, db dbapi.Db) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if !cfg.Authentication.Local.Enabled {
			http.NotFound(resp, req)
			return
		}
		if err := req.ParseForm(); err != nil {
			slog.Error("Failed to parse local login form", "error", err)
			http.Error(resp, "Bad Request", http.StatusBadRequest)
			return
		}
		username := req.Form.Get("username")
		password := req.Form.Get("password")
		if username == "" || password == "" {
			http.Error(resp, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		ctx := req.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			slog.Error("Failed to open database transaction", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		cred, err := tx.Query().GetUserByUsername(ctx, username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Equalize response time with a real verification so that
				// username existence cannot be inferred from timing.
				_, _ = verifyPassword(password, sql.NullString{Valid: true, String: dummyHash})
				http.Error(resp, "Invalid credentials", http.StatusUnauthorized)
				return
			}
			slog.Error("Failed to get user by username", "username", username, "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		matched, err := verifyPassword(password, cred.PasswordHash)
		if err != nil {
			slog.Error("Failed to verify password", "username", cred.Username, "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !matched {
			http.Error(resp, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := generateSessionToken()
		if err != nil {
			slog.Error("Failed to generate session token", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		expiration := time.Now().Add(24 * time.Hour)
		tokenHash := sha256.Sum256([]byte(token))

		if err := tx.Query().CreateUserSession(ctx, queries.CreateUserSessionParams{
			TokenHash: tokenHash[:],
			UserID:    cred.ID,
			Expires:   expiration,
		}); err != nil {
			slog.Error("Failed to create user session", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			slog.Error("Failed to commit transaction", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(resp, &http.Cookie{
			Name:     "shaunkeyscom-session",
			Value:    token,
			Expires:  expiration,
			HttpOnly: true,
			Secure:   !env.Devmode,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		})
		http.Redirect(resp, req, "/", http.StatusSeeOther)
	})
}

// hashPassword derives a password Argon2id hash and returns it in PHC string
// format (self-describing, so parameters can evolve over time).
func hashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2Params)
}

// verifyPassword checks a password against a stored PHC-format Argon2id hash
// in constant time. An empty stored hash never matches.
func verifyPassword(password string, encoded sql.NullString) (bool, error) {
	if !encoded.Valid || encoded.String == "" {
		return false, nil
	}
	return argon2id.ComparePasswordAndHash(password, encoded.String)
}

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
		if err := tx.Query().DeleteUserSession(req.Context(), userSession.TokenHash); err != nil {
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
