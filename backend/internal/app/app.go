package app

import (
	"log/slog"

	"github.com/kkonst40/cloud-storage/backend/internal/api"
	authhandler "github.com/kkonst40/cloud-storage/backend/internal/api/handler/auth"
	profilehandler "github.com/kkonst40/cloud-storage/backend/internal/api/handler/profile"
	resourcehandler "github.com/kkonst40/cloud-storage/backend/internal/api/handler/resource"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
	"github.com/kkonst40/cloud-storage/backend/internal/service/jwt"
	"github.com/kkonst40/cloud-storage/backend/internal/service/s3"
	sessionservice "github.com/kkonst40/cloud-storage/backend/internal/service/session"
	userservice "github.com/kkonst40/cloud-storage/backend/internal/service/user"
	"github.com/kkonst40/cloud-storage/backend/internal/storage"
	userrepo "github.com/kkonst40/cloud-storage/backend/internal/storage/user"
)

const pkg = "App"

type App struct {
	apiClient *api.Client
	dbClient  *storage.Client
}

func New(cfg *config.Config) (*App, error) {
	dbClient, err := storage.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	s3Client, err := s3.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	redisClient, err := sessionservice.NewRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	var (
		userRepo = userrepo.NewRepository(dbClient.DB())
	)

	var (
		userService    = userservice.New(userRepo)
		jwtService     = jwt.New(cfg)
		sessionService = sessionservice.New(cfg, redisClient, jwtService)
		s3Service      = s3.NewService(cfg, s3Client)
	)

	var (
		profileHandler  = profilehandler.New(userService)
		authHandler     = authhandler.New(cfg, userService, sessionService)
		resourceHandler = resourcehandler.New(s3Service)
	)

	apiClient := api.NewClient(cfg, authHandler, resourceHandler, profileHandler, sessionService)

	return &App{
		apiClient: apiClient,
		dbClient:  dbClient,
	}, nil
}

func (a *App) Run() {
	a.apiClient.Start()
}

func (a *App) Shutdown() {
	const op = "Shutdown"

	if err := a.apiClient.Shutdown(); err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)
	}

	if err := a.dbClient.Shutdown(); err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)
	}
}
