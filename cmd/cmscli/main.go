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

	var protocols http.Protocols
	protocols.SetUnencryptedHTTP2(true)
	httpClient := http.Client{
		Transport: &http.Transport{
			Protocols: &protocols,
		},
	}

	client := cmsv1connect.NewAuthServiceClient(&httpClient, "http://127.0.0.1:8080")

	authKey, err := Authenticate(ctx, client)
	if err != nil {
		panic(err)
	}
	slog.Info("Got auth key", "key", authKey)
}

func Authenticate(ctx context.Context, client cmsv1connect.AuthServiceClient) (string, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return "", ErrNoKeys
	}

	conn, err := net.Dial("unix", sock)
	if err != nil {
		return "", fmt.Errorf("failed to connect to ssh agent: %w", err)
	}

	myAgent := agent.NewClient(conn)

	keys, err := myAgent.List()
	if err != nil {
		return "", fmt.Errorf("failed to list ssh keys: %w", err)
	}

	exchanger, err := client.ExchangeSSHKey(ctx)
	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		fingerprint := ssh.FingerprintSHA256(key)

		// Check if fingerprint is acceptable
		if err := exchanger.Send(cmsv1.ExchangeSSHKeyRequest_builder{
			Fingerprint: cmsv1.ExchangeSSHKeyRequest_Fingerprint_builder{
				Fingerprint: fingerprint,
			}.Build(),
		}.Build()); err != nil {
			return "", err
		}
		response, err := exchanger.Receive()
		if err != nil {
			return "", err
		}

		switch {
		case response.HasApproved():
			approvedMsg := response.GetApproved()
			challenge := approvedMsg.GetNonce()
			signed, err := myAgent.Sign(key, challenge)
			if err != nil {
				return "", err
			}
			if err := exchanger.Send(cmsv1.ExchangeSSHKeyRequest_builder{
				Signature: cmsv1.ExchangeSSHKeyRequest_Signature_builder{
					Format: signed.Format,
					Signature: signed.Blob,
				}.Build(),
			}.Build()); err != nil {
				return "", err
			}
			response, err := exchanger.Receive()
			if err != nil {
				return "", err
			}
			if !response.HasToken() {
				return "", ErrProtocolViolation
			}
			return response.GetToken().GetAuthToken(), nil
		case response.HasRejected():
			continue
		default:
			slog.Error("Unrecognized message during SSH key authentication", "msg", response)
			return "", ErrProtocolViolation
		}
	}

	return "", errors.New("no keys accepted")
}
