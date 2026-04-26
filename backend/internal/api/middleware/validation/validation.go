package validation

import (
	"regexp"

	"github.com/gofiber/fiber/v3"
	"github.com/kkonst40/cloud-storage/backend/internal/api/dto"
	"github.com/kkonst40/cloud-storage/backend/internal/api/dto/profile"
	"github.com/kkonst40/cloud-storage/backend/internal/api/handler"
)

func CredentialsValidation(ctx fiber.Ctx) error {
	var r profile.LoginRequest
	err := ctx.Bind().Body(&r)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	if !(isValidName(r.Username) && isValidPassword(r.Password)) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(
			&dto.ErrorResponse{Message: handler.MessageLoginOrPasswordInvalid},
		)
	}

	ctx.Locals("user_data", r)

	return ctx.Next()
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]*$`)
var passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()\-=_+]+$`)

func isValidName(name string) bool {
	if len(name) < 5 || 20 < len(name) {
		return false
	}

	return usernameRegex.MatchString(name)
}

func isValidPassword(password string) bool {
	if len(password) < 8 || 72 < len(password) {
		return false
	}

	return passwordRegex.MatchString(password)
}
