package server

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenLifetime = time.Hour * 24 * 365 * 10
)

type jwtAuth struct {
	issuer   string
	audience string
	key      []byte
}

func processKey(input string) []byte {
	key := sha256.Sum256([]byte(input))
	return key[:]
}

func newJWTAuth(cfg *ServerConfig) *jwtAuth {
	return &jwtAuth{
		issuer:   cfg.MyURL,
		audience: cfg.MyURL,
		key:      processKey(cfg.MyKeyInput),
	}
}

func (ja *jwtAuth) keyFunc(_ *jwt.Token) (any, error) {
	return ja.key, nil
}

func (ja *jwtAuth) mintToken(userID, keyInput string) (*Token, error) {
	gotKey := processKey(keyInput)
	if !bytes.Equal(gotKey, ja.key) {
		return nil, errors.New("invalid key")
	}
	now := time.Now()
	expires := now.Add(tokenLifetime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    ja.issuer,
		Subject:   userID,
		Audience:  []string{ja.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expires),
	})
	tokenStr, err := token.SignedString(ja.key)
	if err != nil {
		return nil, fmt.Errorf("signing token: %w", err)
	}
	tpb := &Token{
		RawToken:  tokenStr,
		Issuer:    ja.issuer,
		Subject:   userID,
		Audience:  []string{ja.audience},
		IssuedAt:  now.UnixMilli(),
		ExpiresAt: expires.UnixMilli(),
	}
	return tpb, nil
}

func (ja *jwtAuth) parse(tokenString string) (string, error) {
	t, err := jwt.Parse(tokenString, ja.keyFunc,
		jwt.WithAudience(ja.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(ja.issuer),
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return "", err
	}
	sub, err := t.Claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("getting subject: %w", err)
	}
	if sub == "" {
		return "", errors.New("no subject")
	}
	return sub, nil
}
