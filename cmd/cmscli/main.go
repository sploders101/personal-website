package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"charm.land/log/v2"
	"connectrpc.com/connect"
	"github.com/sploders101/personal-website/internal/authclient"
	"github.com/sploders101/personal-website/internal/env"
	cmsv1 "github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1"
	"github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1/cmsv1connect"
)

func init() {
	handler := log.New(os.Stderr)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// TODO: Source these from other means
const ENDPOINT = "http://127.0.0.1:8080"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	authenticator := authclient.NewAuthenticator()

	client := cmsv1connect.NewCmsServiceClient(
		env.ConnectHttp,
		ENDPOINT,
		connect.WithInterceptors(authenticator),
	)
	resp, err := client.Ping(ctx, cmsv1.PingRequest_builder{Message: "Hello world!"}.Build())
	if err != nil {
		panic(err)
	}
	slog.Info("Got ping response from server!", "message", resp.GetMessage())
}
