package apiservices

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/sploders101/personal-website/cmd/webserver/config"
	"github.com/sploders101/personal-website/cmd/webserver/dbapi"
	"github.com/sploders101/personal-website/internal/authutils"
	cmsv1 "github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1"
	"golang.org/x/crypto/ssh"
)

const (
	MAX_KEY_OFFERINGS = 15
	NONCE_SIZE        = 256
)

type AuthService struct {
	config config.ServerConfig
	db     dbapi.Db
}

func NewAuthService(config config.ServerConfig, db dbapi.Db) AuthService {
	return AuthService{
		config: config,
		db:     db,
	}
}

func (auth AuthService) ExchangeSSHKey(
	ctx context.Context,
	stream *connect.BidiStream[cmsv1.ExchangeSSHKeyRequest, cmsv1.ExchangeSSHKeyResponse],
) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	finished := make(chan error, 1)
	go func() {
		for range MAX_KEY_OFFERINGS {
			offeredKey, err := stream.Receive()
			if err != nil {
				slog.Error("Failed to receive key offering", "error", err)
				finished <- ErrAmbiguousInternal
				return
			}
			if !offeredKey.HasFingerprint() {
				finished <- ErrProtocolViolation
				return
			}
			fingerprintMsg := offeredKey.GetFingerprint()
			fingerprint := fingerprintMsg.GetFingerprint()

			tx, err := auth.db.Begin(ctx)
			if err != nil {
				slog.Error("Failed to open db transaction", "error", err)
				finished <- ErrAmbiguousInternal
				return
			}

			userBundle, err := tx.Query().GetUserBySshKey(ctx, fingerprint)
			tx.Rollback()
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					// Expected. Send rejection
					if err := stream.Send(cmsv1.ExchangeSSHKeyResponse_builder{
						Rejected: cmsv1.ExchangeSSHKeyResponse_FingerprintRejected_builder{}.Build(),
					}.Build()); err != nil {
						return
					}
					continue
				}
				slog.Error("Failed to get user by SSH key", "error", err)
				finished <- ErrAmbiguousInternal
				return
			}

			// Key was found. Start exchange.
			nonce := make([]byte, NONCE_SIZE)
			rand.Read(nonce)
			if err := stream.Send(cmsv1.ExchangeSSHKeyResponse_builder{
				Approved: cmsv1.ExchangeSSHKeyResponse_FingerprintApproved_builder{
					Nonce: nonce,
				}.Build(),
			}.Build()); err != nil {
				return
			}

			signatureContainer, err := stream.Receive()
			if err != nil {
				return
			}
			if !signatureContainer.HasSignature() {
				// TODO: Allow sending another fingerprint instead.
				finished <- ErrProtocolViolation
				return
			}
			signatureMsg := signatureContainer.GetSignature()
			sigFmt := signatureMsg.GetFormat()
			signature := signatureMsg.GetSignature()

			key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(userBundle.UsersSshKey.PublicKey))
			if err != nil {
				finished <- ErrProtocolViolation
				return
			}
			if err := key.Verify(nonce, &ssh.Signature{
				Format: sigFmt,
				Blob:   signature,
			}); err != nil {
				slog.Error("Error validating signature", "error", err)
				finished <- ErrAuthenticationFailed
				return
			}

			// User is authenticated. Send them a key.
			claims := authutils.IdentityClaims{
				SshKeyFingerprint: fingerprint,
			}
			claims.FillRegistered(userBundle.User.Username, 15*time.Minute)
			tokenString, err := authutils.SignToken(claims, auth.config.Secrets.JwtSecret)
			if err != nil {
				slog.Error("Failed to sign JWT", "error", err)
				finished <- ErrAmbiguousInternal
				return
			}
			if err := stream.Send(cmsv1.ExchangeSSHKeyResponse_builder{
				Token: cmsv1.ExchangeSSHKeyResponse_AuthToken_builder{
					AuthToken: tokenString,
				}.Build(),
			}.Build()); err != nil {
				finished <- ErrAmbiguousInternal
				return
			}
			finished <- nil
			return
		}
	}()

	select {
	case res := <-finished:
		return res
	case <-ctx.Done():
		return ErrTimedOut
	}
}
