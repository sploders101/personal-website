package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sploders101/personal-website/internal/config"
	"golang.org/x/oauth2"
)

func registerOidcHandlers(ctx context.Context, cfg config.ServerConfig, mux *http.ServeMux) error {
	for providerName, providerCfg := range cfg.Authentication.Oidc {
		handler, err := createHandler(ctx, cfg.BaseUrl, providerName, providerCfg)
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
	provider  *oidc.Provider
	oauth2Cfg *oauth2.Config
	verifier  *oidc.IDTokenVerifier
}

func (handler *OIDCHandler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/oidc/"+handler.name+"/begin", handler.handleBegin)
	mux.HandleFunc("GET /auth/oidc/"+handler.name+"/callback", handler.handleCallback)
}

func (handler *OIDCHandler) handleBegin(resp http.ResponseWriter, req *http.Request) {
	state, err := generateState()
	if err != nil {
		slog.Error("Failed to generate state", "error", err.Error())
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

	token, err := handler.oauth2Cfg.Exchange(req.Context(), code)
	if err != nil {
		slog.Error("Failed to exchange code for token", "error", err)
		http.Error(resp, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		slog.Error("No id_token in response")
		http.Error(resp, "No id_token in response", http.StatusInternalServerError)
		return
	}

	idToken, err := handler.verifier.Verify(ctx, rawIdToken)
	if err != nil {
		slog.Error("Failed to verify ID token", "error", err)
		http.Error(resp, "Failed to verify ID token", http.StatusInternalServerError)
		return
	}

	var claims struct {
		Sub string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		slog.Error("Failed to parse claims", "error", err)
		http.Error(resp, "Failed to parse claims", http.StatusInternalServerError)
		return
	}

	userInfo, err := handler.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		slog.Error("Failed to hit userinfo endpoint", "provider", handler.name, "sub", claims.Sub, "error", err)
		http.Error(resp, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	var uiClaims UserInfoClaims
	if err := userInfo.Claims(&uiClaims); err != nil {
		slog.Error("Failed to parse userinfo claims", "provider", handler.name, "sub", claims.Sub, "error", err)
		http.Error(resp, "Failed to parse user info", http.StatusInternalServerError)
		return
	}
	resp.Write([]byte("Authentication successful! You are " + uiClaims.Username))
}

func createHandler(ctx context.Context, baseUrl string, providerName string, cfg config.AuthenticationOidcConfig) (*OIDCHandler, error) {
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
		providerName,
		provider,
		oauth2Cfg,
		verifier,
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
