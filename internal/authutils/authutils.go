package authutils

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type IdentityClaims struct {
	jwt.RegisteredClaims

	SshKeyFingerprint string `json:"ssh_key_fingerprint"`
}

func (claims IdentityClaims) FillRegistered(subject string, duration time.Duration) {
	now := time.Now()
	claims.Issuer = "shaunkeyscom-cms-auth"
	claims.Subject = subject
	claims.Audience = []string{"shaunkeyscom-cms"}
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(duration))
	claims.NotBefore = jwt.NewNumericDate(now)
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ID = rand.Text()
}

func SignToken(claims IdentityClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string, secret string) (IdentityClaims, error) {
	var claims IdentityClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid algorithm")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return IdentityClaims{}, err
	}
	return claims, nil
}
