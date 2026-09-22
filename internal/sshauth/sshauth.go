package sshauth

import (
	"errors"
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var ErrNoKeys = errors.New("no valid ssh keys found")

func SignSomething(msg string) error {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return ErrNoKeys
	}

	conn, err := net.Dial("unix", sock)
	if err != nil {
		return fmt.Errorf("failed to connect to ssh agent: %w", err)
	}

	agent := agent.NewClient(conn)
	agent.List()
}
