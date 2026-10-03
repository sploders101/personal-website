package main

import (
	"context"
	"log/slog"
	"os"

	"charm.land/log/v2"
	"connectrpc.com/connect"
	"github.com/sploders101/personal-website/internal/authclient"
	"github.com/sploders101/personal-website/internal/env"
	"github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1/cmsv1connect"
	"github.com/urfave/cli/v3"
)

// TODO: Source these from other means
const ENDPOINT = "http://127.0.0.1:8080"
const ASSET_MAX_CHUNK_SIZE = 8192

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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Aliases: []string{"m"},
						Name:    "message",
						Value:   "Hello, world!",
						Usage:   "Changes the message the server should send back.",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return pingAPI(ctx, cmd.String("message"))
				},
			},
			{
				Name:  "push",
				Usage: "Pushes an article to the CMS.",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "mdfile",
						Required:  true,
						UsageText: "<mdfile>",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return pushArticle(ctx, cmd.StringArg("mdfile"))
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error())
	}
}

func getClient() cmsv1connect.CmsServiceClient {
	authenticator := authclient.NewAuthenticator(ENDPOINT)
	return cmsv1connect.NewCmsServiceClient(
		env.ConnectHttp,
		ENDPOINT,
		connect.WithInterceptors(authenticator),
	)
}
