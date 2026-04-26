package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kkonst40/cloud-storage/backend/internal/app"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
)

func main() {
	cfg := config.MustNew("")
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	appCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	application.Run()

	<-appCtx.Done()

	_, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	application.Shutdown()
}
