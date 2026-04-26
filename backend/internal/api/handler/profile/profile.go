package profile

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/kkonst40/cloud-storage/backend/internal/api/dto"
	"github.com/kkonst40/cloud-storage/backend/internal/api/dto/profile"
	"github.com/kkonst40/cloud-storage/backend/internal/api/handler"
	"github.com/kkonst40/cloud-storage/backend/internal/domain"
	userservice "github.com/kkonst40/cloud-storage/backend/internal/service/user"
)

const pkg = "ProfileHandler"

type Handler struct {
	userService UserService
}

type UserService interface {
	UserById(userId int64) (domain.User, error)
}

func New(userService UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

func (p *Handler) Show(ctx fiber.Ctx) error {
	const op = "Show"

	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)

	user, err := p.userService.UserById(userId)
	if err != nil {
		if !errors.Is(err, userservice.ErrNotFound) {
			slog.Error(err.Error(), "pkg", pkg, "op", op)
		}

		return ctx.Status(fiber.StatusUnauthorized).JSON(&dto.ErrorResponse{Message: handler.MessageUnauthorized})
	}

	ctx.Status(fiber.StatusOK)

	return ctx.JSON(&profile.ProfileResponse{
		Username: user.Username,
	})
}
