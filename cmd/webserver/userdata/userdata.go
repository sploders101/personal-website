package userdata

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sploders101/personal-website/internal/dbapi"
	queries "github.com/sploders101/personal-website/internal/dbapi/gen"
)

type userInfoKey struct{}
type sessionInfoKey struct{}

func UserMiddleware(db dbapi.Db, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		sessionCookie, err := req.Cookie("shaunkeyscom-session")
		if err != nil {
			handler.ServeHTTP(resp, req)
			return
		}

		// Since there's a cookie, validate it
		tx, err := db.Begin(req.Context())
		if err != nil {
			slog.Error("Unable to open database transaction", "error", err)
			handler.ServeHTTP(resp, req)
			return
		}
		tokenHash := sha256.Sum256([]byte(sessionCookie.Value))
		session, err := tx.Query().GetSessionDataByTokenHash(req.Context(), tokenHash[:])
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				slog.Error("Unable to get session data", "error", err)
			}
			handler.ServeHTTP(resp, req)
			return
		}
		tx.Rollback()
		ctx := context.WithValue(req.Context(), userInfoKey{}, session.User)
		ctx = context.WithValue(ctx, sessionInfoKey{}, session.UsersSession)
		req = req.WithContext(ctx)
		handler.ServeHTTP(resp, req)
	})
}

// Gets the user data from the request.
func GetUserData(ctx context.Context) queries.User {
	val := ctx.Value(userInfoKey{})
	if val == nil {
		return queries.User{}
	}
	return val.(queries.User)
}

// Gets the session information from the request
func GetSessionInfo(ctx context.Context) queries.UsersSession {
	val := ctx.Value(sessionInfoKey{})
	if val == nil {
		return queries.UsersSession{}
	}
	return val.(queries.UsersSession)
}
