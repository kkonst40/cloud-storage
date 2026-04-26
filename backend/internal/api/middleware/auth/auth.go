package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/kkonst40/cloud-storage/backend/internal/api/dto"
	"github.com/kkonst40/cloud-storage/backend/internal/api/handler"
	"github.com/kkonst40/cloud-storage/backend/internal/config"

	"github.com/gofiber/fiber/v3"
)

type SessionService interface {
	ValidateAccessToken(accessTokenStr string) (int64, error)
	RefreshSession(ctx context.Context, refreshTokenStr string) (string, string, error)
}

func Auth(cfg *config.Config, sessionService SessionService) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		accessTokenStr := ctx.Cookies("access_token")

		userId, err := sessionService.ValidateAccessToken(accessTokenStr)
		if err == nil {
			ctx.Locals("user_id", userId)
			return ctx.Next()
		}

		refreshTokenStr := ctx.Cookies("refresh_token")
		if refreshTokenStr == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(
				&dto.ErrorResponse{Message: handler.MessageUnauthorized},
			)
		}

		newAccessToken, newRefreshToken, err := sessionService.RefreshSession(ctx, refreshTokenStr)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(
				&dto.ErrorResponse{Message: handler.MessageUnauthorized},
			)
		}

		accessTokenExpiresAt := time.Now().Add(time.Duration(cfg.AccessTokenExpiresMinutes) * time.Minute)
		refreshTokenExpiresAt := time.Now().Add(time.Duration(cfg.RefreshTokenExpiresHours) * time.Hour)

		ctx.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			Path:     "/",
			HTTPOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: cfg.CookieSameSite,
			Expires:  accessTokenExpiresAt,
		})

		ctx.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			Path:     "/",
			HTTPOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: cfg.CookieSameSite,
			Expires:  refreshTokenExpiresAt,
		})

		userId, err = sessionService.ValidateAccessToken(newAccessToken)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				&dto.ErrorResponse{Message: handler.MessageUnauthorized},
			)
		}
		slog.Debug("Auth middleware", "userID", userId, "access", newAccessToken, "refresh", newRefreshToken)

		ctx.Locals("user_id", userId)
		return ctx.Next()
	}
}
