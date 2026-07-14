package jwt

import (
	"testing"

	"github.com/kkonst40/cloud-storage/backend/internal/config"
)

func TestService_GenerateAccessToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:                 "test_secret",
		JWTIssuer:                 "test_issuer",
		AccessTokenExpiresMinutes: 1,
	}

	jwtService := New(cfg)
	userId := int64(123456789)
	token, err := jwtService.GenerateAccessToken(userId)
	if err != nil {
		t.Error("generate access token error", err)
	}

	if token == "" {
		t.Error("empty access token")
	}
}

func TestService_ValidateAccessToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:                 "test_secret",
		JWTIssuer:                 "test_issuer",
		AccessTokenExpiresMinutes: 1,
	}

	jwtService := New(cfg)
	userId := int64(123456789)
	token, err := jwtService.GenerateAccessToken(userId)
	if err != nil {
		t.Error("generate access token error", err)
	}

	if token == "" {
		t.Error("empty access token")
	}

	outUserId, err := jwtService.ValidateAccessToken(token)
	if err != nil {
		t.Error("validate access token error", err)
	}

	if userId != outUserId {
		t.Errorf("subject must be %v, got: %v", userId, outUserId)
	}
}
