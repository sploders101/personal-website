package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"connectrpc.com/connect"
	cmsv1 "github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1"
	"github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1/cmsv1connect"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var ErrNoKeys = errors.New("no valid ssh keys found")
var ErrProtocolViolation = errors.New("protocol violation")

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client := cmsv1connect.NewAuthServiceClient(http.DefaultClient, "http://127.0.0.1:8080")

	if err := SignSomething(ctx, client); err != nil {
		panic(err)
	}
}

func SignSomething(ctx context.Context, client cmsv1connect.AuthServiceClient) error {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return ErrNoKeys
	}

	conn, err := net.Dial("unix", sock)
	if err != nil {
		return fmt.Errorf("failed to connect to ssh agent: %w", err)
	}

	myAgent := agent.NewClient(conn)

	keys, err := myAgent.List()
	if err != nil {
		return fmt.Errorf("failed to list ssh keys: %w", err)
	}

	exchanger, err := client.ExchangeSSHKey(ctx)
	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		fingerprint := ssh.FingerprintSHA256(key)

		// Check if fingerprint is acceptable
		exchanger.Send(cmsv1.ExchangeSSHKeyRequest_builder{
			Fingerprint: cmsv1.ExchangeSSHKeyRequest_Fingerprint_builder{
				Fingerprint: fingerprint,
			}.Build(),
		}.Build())
		response, err := exchanger.Receive()
		if err != nil {
			return err
		}

		switch {
		case response.HasApproved():
			approvedMsg := response.GetApproved()
			return exchangeKey(myAgent, exchanger, key, approvedMsg.GetNonce())
		case response.HasRejected():
			continue
		default:
			slog.Error("Unrecognized message during SSH key authentication", "msg", response)
			return ErrProtocolViolation
		}
	}

	return nil
}

func exchangeKey(
	myAgent agent.Agent,
	exchanger *connect.BidiStreamForClientSimple[cmsv1.ExchangeSSHKeyRequest, cmsv1.ExchangeSSHKeyResponse],
	key *agent.Key,
	challenge []byte,
) error {
	
}
