package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrUnexpectedMethod = errors.New("unexpected signing method")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidClaims    = errors.New("invalid claims")
)

type Service struct {
	secret         []byte
	issuer         string
	accessTokenTTL time.Duration
}

func New(cfg *config.Config) *Service {
	return &Service{
		secret:         []byte(cfg.JWTSecret),
		issuer:         cfg.JWTIssuer,
		accessTokenTTL: time.Duration(cfg.AccessTokenExpiresMinutes) * time.Minute,
	}
}

func (s *Service) GenerateAccessToken(userId int64) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTokenTTL)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userId, 10),
		Issuer:    s.issuer,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.secret)
}

func (s *Service) ValidateAccessToken(tokenStr string) (int64, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: %v", ErrUnexpectedMethod, token.Header["alg"])
			}
			return s.secret, nil
		},
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return 0, ErrInvalidToken
		}
		return 0, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !(ok && token.Valid) {
		return 0, ErrInvalidClaims
	}

	if claims.Subject == "" {
		return 0, ErrInvalidClaims
	}
	if s.issuer != "" && claims.Issuer != s.issuer {
		return 0, ErrInvalidClaims
	}

	userId, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, ErrInvalidClaims
	}

	return userId, nil
}
