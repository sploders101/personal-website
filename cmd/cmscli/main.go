package main

import (
	"context"
	"log/slog"
	"os"

	"charm.land/log/v2"
	"connectrpc.com/connect"
	"github.com/sploders101/personal-website/internal/authclient"
	"github.com/sploders101/personal-website/internal/env"
	cmsv1 "github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1"
	"github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1/cmsv1connect"
	"github.com/sploders101/personal-website/internal/markdown"
	"github.com/urfave/cli/v3"
)

// TODO: Source these from other means
const ENDPOINT = "http://127.0.0.1:8080"

func main() {
	handler := log.New(os.Stderr)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if env.Devmode {
		slog.Warn("Devmode is enabled!")
	}

	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "ping",
				Usage: "Authenticate against the CMS and ping the API.",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "message",
						Value:     "Hello, world!",
						UsageText: "Changes the message the server should send back.",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return pingAPI(ctx, cmd.StringArg("message"))
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error())
	}
}

func pingAPI(ctx context.Context, message string) error {
	authenticator := authclient.NewAuthenticator()

	client := cmsv1connect.NewCmsServiceClient(
		env.ConnectHttp,
		ENDPOINT,
		connect.WithInterceptors(authenticator),
	)
	resp, err := client.Ping(ctx, cmsv1.PingRequest_builder{Message: message}.Build())
	if err != nil {
		return err
	}
	slog.Info(
		"Got ping response from server!",
		"message",
		resp.GetMessage(),
		"userID",
		resp.GetUserId(),
	)

	return nil
}
