package jwt

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kkonst40/cloud-storage/backend/internal/config"
)

func TestJWT_GenerateAccessToken(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Fatal(err)
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

func TestJWT_ValidateAccessToken(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Fatal(err)
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

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func getConfig() (*config.Config, error) {
	dir, err := findProjectRoot()
	if err != nil {
		return nil, err
	}

	return config.MustNew(filepath.Join(dir, ".env.dev")), nil
}
