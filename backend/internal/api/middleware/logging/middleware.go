package logging

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Logging(ctx fiber.Ctx) error {
	start := time.Now()

	method := ctx.Method()
	url := ctx.FullURL()

	err := ctx.Next()

	duration := time.Since(start)
	statusCode := ctx.Response().StatusCode()

	logAttrs := []any{
		slog.String("method", method),
		slog.String("path", url),
		slog.Int("status", statusCode),
		slog.Duration("duration", duration),
	}

	if err != nil {
		logAttrs = append(logAttrs, slog.String("error", err.Error()))
		slog.Error("HTTP request error", logAttrs...)
	} else {
		slog.Info("HTTP request processed", logAttrs...)
	}

	return err
}
