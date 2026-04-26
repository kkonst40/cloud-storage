package session

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
	errs "github.com/kkonst40/cloud-storage/backend/internal/errors"
	"github.com/redis/go-redis/v9"
)

const pkg = "SessionService"

var (
	ErrNotFound       = errors.New("user session not found")
	ErrAlreadyExists  = errors.New("user session already exists")
	ErrSessionExpired = errors.New("user session expired")
)

type Service struct {
	redisClient *redis.Client
	jwtService  JwtService
	tokenTTL    time.Duration
}

type JwtService interface {
	GenerateAccessToken(userId int64) (string, error)
	ValidateAccessToken(token string) (int64, error)
}

func New(cfg *config.Config, redisClient *redis.Client, jwtService JwtService) *Service {
	return &Service{
		redisClient: redisClient,
		jwtService:  jwtService,
		tokenTTL:    time.Duration(cfg.RefreshTokenExpiresHours) * time.Hour,
	}
}

func (s *Service) RefreshSession(ctx context.Context, refreshToken string) (string, string, error) {
	const op = "RefreshSession"

	userIdStr, err := s.redisClient.Get(ctx, "session:"+refreshToken).Result()
	if err != nil {
		return "", "", errs.Wrap(pkg, op, err)
	}

	if err := s.redisClient.Del(ctx, "session:"+refreshToken).Err(); err != nil {
		return "", "", errs.Wrap(pkg, op, err)
	}

	newRefreshToken := uuid.New().String()
	if err := s.redisClient.Set(ctx, "session:"+newRefreshToken, userIdStr, s.tokenTTL).Err(); err != nil {
		return "", "", errs.Wrap(pkg, op, err)
	}

	userId, _ := strconv.ParseInt(userIdStr, 10, 64)
	newAccessToken, err := s.jwtService.GenerateAccessToken(userId)
	if err != nil {
		return "", "", errs.Wrap(pkg, op, err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *Service) ValidateAccessToken(accessToken string) (int64, error) {
	return s.jwtService.ValidateAccessToken(accessToken)
}

func (s *Service) CreateSession(ctx context.Context, userId int64) (string, string, error) {
	const op = "CreateSession"

	refreshToken := uuid.New().String()
	if err := s.redisClient.Set(ctx, "session:"+refreshToken, fmt.Sprintf("%d", userId), s.tokenTTL).Err(); err != nil {
		return "", "", errs.Wrap(pkg, op, err)
	}

	accessToken, err := s.jwtService.GenerateAccessToken(userId)
	if err != nil {
		return "", "", errs.Wrap(pkg, op, err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) DeleteSession(ctx context.Context, userId int64, refreshToken string) error {
	const op = "DeleteSession"

	if err := s.redisClient.Del(ctx, "session:"+refreshToken).Err(); err != nil {
		return errs.Wrap(pkg, op, err)
	}

	return nil
}
