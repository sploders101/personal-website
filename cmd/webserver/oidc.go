package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sploders101/personal-website/internal/config"
	"github.com/sploders101/personal-website/internal/dbapi"
	queries "github.com/sploders101/personal-website/internal/dbapi/gen"
	"github.com/sploders101/personal-website/internal/env"
	"golang.org/x/oauth2"
)

func registerOidcHandlers(
	ctx context.Context,
	cfg config.ServerConfig,
	db dbapi.Db,
	mux *http.ServeMux,
) error {
	for providerName, providerCfg := range cfg.Authentication.Oidc {
		handler, err := createHandler(ctx, db, cfg.BaseUrl, providerName, providerCfg)
		if err != nil {
			// TODO: Fail safe. No auth means no admin panel. Content can still be served.
			return err
		}
		handler.registerRoutes(mux)
	}

	return nil
}

type OIDCHandler struct {
	name      string
	config    config.AuthenticationOidcConfig
	db        dbapi.Db
	provider  *oidc.Provider
	oauth2Cfg *oauth2.Config
	verifier  *oidc.IDTokenVerifier
}

func (handler *OIDCHandler) validateOidc(
	ctx context.Context,
	state string,
	code string,
) (UserInfoClaims, error) {
	token, err := handler.oauth2Cfg.Exchange(ctx, code)
	if err != nil {
		slog.Error("Failed to exchange code for token", "error", err)
		return UserInfoClaims{}, err
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		slog.Error("No id_token in response")
		return UserInfoClaims{}, err
	}

	idToken, err := handler.verifier.Verify(ctx, rawIdToken)
	if err != nil {
		slog.Error("Failed to verify ID token", "error", err)
		return UserInfoClaims{}, err
	}

	var claims struct {
		Sub string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		slog.Error("Failed to parse claims", "error", err)
		return UserInfoClaims{}, err
	}

	userInfo, err := handler.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		slog.Error(
			"Failed to hit userinfo endpoint",
			"provider",
			handler.name,
			"sub",
			claims.Sub,
			"error",
			err,
		)
		return UserInfoClaims{}, err
	}

	var uiClaims UserInfoClaims
	if err := userInfo.Claims(&uiClaims); err != nil {
		slog.Error(
			"Failed to parse userinfo claims",
			"provider",
			handler.name,
			"sub",
			claims.Sub,
			"error",
			err,
		)
		return UserInfoClaims{}, err
	}

	return uiClaims, nil
}

func (handler *OIDCHandler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/oidc/"+handler.name+"/begin", handler.handleBegin)
	mux.HandleFunc("GET /auth/oidc/"+handler.name+"/callback", handler.handleCallback)
}

func (handler *OIDCHandler) handleBegin(resp http.ResponseWriter, req *http.Request) {
	state, err := generateState()
	if err != nil {
		slog.Error("Failed to generate state", "error", err)
		http.Error(resp, "Internal server error", http.StatusInternalServerError)
		return
	}

	authUrl := handler.oauth2Cfg.AuthCodeURL(state)
	http.Redirect(resp, req, authUrl, http.StatusFound)
}

type UserInfoClaims struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Username      string `json:"preferred_username"`
}

func (handler *OIDCHandler) handleCallback(resp http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	state := req.URL.Query().Get("state")
	if state == "" {
		http.Error(resp, "Missing state parameter", http.StatusBadRequest)
		return
	}

	code := req.URL.Query().Get("code")
	if code == "" {
		http.Error(resp, "Missing code parameter", http.StatusBadRequest)
		return
	}

	claims, err := handler.validateOidc(ctx, state, code)
	if err != nil {
		// Error already logged
		http.Error(resp, "Unable to validate credential", http.StatusUnauthorized)
	}

	// Create user account

	tx, err := handler.db.Begin(ctx)
	if err != nil {
		slog.Error("Failed to start transaction", "error", err)
		http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	user, err := tx.Query().GetUserByOidc(
		ctx,
		queries.GetUserByOidcParams{
			Issuer:  handler.name,
			Subject: claims.Subject,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) && handler.config.AllowRegistration {
			username := claims.Subject
			if claims.Username != "" {
				username = claims.Username
			}
			newUser, err := tx.Query().CreateUser(ctx, queries.CreateUserParams{
				Username: username,
				Email:    claims.Email,
			})
			if err != nil {
				slog.Error("Failed to create user", "sub", claims.Subject, "error", err)
				return
			}
			user.User = newUser
			oidcIdentity, err := tx.Query().CreateOidcIdentity(
				ctx, queries.CreateOidcIdentityParams{
					UserID:  newUser.ID,
					Issuer:  handler.name,
					Subject: claims.Subject,
				},
			)
			if err != nil {
				slog.Error("Failed to create OIDC identity", "sub", claims.Subject, "error", err)
				return
			}
			user.UsersOidcIdentity = oidcIdentity
		} else {
			slog.Error("Failed to get user", "error", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	token, err := generateSessionToken()
	if err != nil {
		slog.Error("Failed to generate session token", "error", err)
		http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	expiration := time.Now().Add(24 * time.Hour)
	tokenHash := sha256.Sum256([]byte(token))

	err = tx.Query().CreateUserSession(ctx, queries.CreateUserSessionParams{
		TokenHash: tokenHash[:],
		UserID:    user.User.ID,
		Expires:   expiration,
	})
	if err != nil {
		slog.Error("Failed to create user session", "error", err)
		http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit changes", "error", err)
		http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(resp, &http.Cookie{
		Name:     "shaunkeyscom-session",
		Value:    token,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   !env.Devmode,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	http.Redirect(resp, req, "/", http.StatusFound)
}

func createHandler(
	ctx context.Context,
	db dbapi.Db,
	baseUrl string,
	providerName string,
	cfg config.AuthenticationOidcConfig,
) (*OIDCHandler, error) {
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	redirectUrl, err := url.JoinPath(baseUrl, "auth", "oidc", providerName, "callback")
	if err != nil {
		return nil, err
	}

	oauth2Cfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectUrl,
		Scopes:       scopes,
	}

	return &OIDCHandler{
		name:      providerName,
		config:    cfg,
		db:        db,
		provider:  provider,
		oauth2Cfg: oauth2Cfg,
		verifier:  verifier,
	}, nil
}

func generateState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

func generateSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}
