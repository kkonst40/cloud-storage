package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/kkonst40/cloud-storage/backend/internal/api/dto"
	"github.com/kkonst40/cloud-storage/backend/internal/api/dto/profile"
	"github.com/kkonst40/cloud-storage/backend/internal/api/handler"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
	"github.com/kkonst40/cloud-storage/backend/internal/domain"
	"github.com/kkonst40/cloud-storage/backend/internal/service/password"
	userservice "github.com/kkonst40/cloud-storage/backend/internal/service/user"

	"github.com/gofiber/fiber/v3"
)

const pkg = "AuthHandler"

type Handler struct {
	userService     UserService
	sessionService  SessionService
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	cookieSameSite  string
	cookieSecure    bool
}

type SessionService interface {
	CreateSession(ctx context.Context, userId int64) (string, string, error)
	DeleteSession(ctx context.Context, userId int64, refreshToken string) error
}

type UserService interface {
	CreateUser(name, pwd string) (domain.User, error)
	UserByName(name string) (domain.User, error)
}

func New(
	cfg *config.Config,
	userService UserService,
	sessionService SessionService,
) *Handler {
	return &Handler{
		userService:     userService,
		sessionService:  sessionService,
		accessTokenTTL:  time.Duration(cfg.AccessTokenExpiresMinutes) * time.Minute,
		refreshTokenTTL: time.Duration(cfg.RefreshTokenExpiresHours) * time.Hour,
		cookieSameSite:  cfg.CookieSameSite,
		cookieSecure:    cfg.CookieSecure,
	}
}

func (h *Handler) Login(ctx fiber.Ctx) error {
	const op = "Login"

	handler.SetCommonHeaders(ctx)
	r := handler.RequestedLogin(ctx)

	user, err := h.userService.UserByName(r.Username)
	if err != nil {
		if !errors.Is(err, userservice.ErrNotFound) {
			slog.Error(err.Error(), "pkg", pkg, "op", op)
		}

		return ctx.Status(fiber.StatusUnauthorized).JSON(
			&dto.ErrorResponse{Message: handler.MessageLoginOrPasswordInvalid},
		)
	}

	if !password.VerifyPwd(r.Password, user.Password) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(
			&dto.ErrorResponse{Message: handler.MessageLoginOrPasswordInvalid},
		)
	}

	accessToken, refreshToken, err := h.sessionService.CreateSession(ctx, user.ID)
	if err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	h.setCookie(ctx, accessToken, "access_token", time.Now().Add(h.accessTokenTTL))
	h.setCookie(ctx, refreshToken, "refresh_token", time.Now().Add(h.refreshTokenTTL))
	ctx.Status(fiber.StatusCreated)

	return ctx.JSON(&profile.LoginResponse{
		Username: user.Username,
	})
}

func (h *Handler) Register(ctx fiber.Ctx) error {
	const op = "Register"

	handler.SetCommonHeaders(ctx)
	r := handler.RequestedLogin(ctx)

	user, err := h.userService.CreateUser(r.Username, r.Password)
	if err != nil {
		if errors.Is(err, userservice.ErrAlreadyExists) {
			return ctx.Status(fiber.StatusConflict).JSON(
				&dto.ErrorResponse{Message: handler.MessageUserAlreadyExists},
			)
		}

		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	accessToken, refreshToken, err := h.sessionService.CreateSession(ctx, user.ID)
	if err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	h.setCookie(ctx, accessToken, "access_token", time.Now().Add(h.accessTokenTTL))
	h.setCookie(ctx, refreshToken, "refresh_token", time.Now().Add(h.refreshTokenTTL))

	ctx.Status(fiber.StatusCreated)

	return ctx.JSON(&profile.RegisterResponse{
		Username: user.Username,
	})
}

func (h *Handler) Logout(ctx fiber.Ctx) error {
	const op = "Logout"

	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)
	refreshToken := ctx.Cookies("refresh_token")

	h.sessionService.DeleteSession(ctx, userId, refreshToken)

	h.setCookie(ctx, "", "access_token", time.Now())
	h.setCookie(ctx, "", "refresh_token", time.Now())
	ctx.Status(fiber.StatusOK)

	return nil
}

func (h *Handler) setCookie(ctx fiber.Ctx, token, tokenName string, expires time.Time) {
	ctx.Cookie(&fiber.Cookie{
		Name:     tokenName,
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.cookieSameSite,
		Expires:  expires,
	})
}
