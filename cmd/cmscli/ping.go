package main

import (
	"context"
	"log/slog"

	cmsv1 "github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1"
)

// Pings the API to see if we can properly authenticate
func pingAPI(ctx context.Context, message string) error {
	client := getClient()

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
