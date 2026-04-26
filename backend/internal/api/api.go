package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	authhandler "github.com/kkonst40/cloud-storage/backend/internal/api/handler/auth"
	profilehandler "github.com/kkonst40/cloud-storage/backend/internal/api/handler/profile"
	resourcehandler "github.com/kkonst40/cloud-storage/backend/internal/api/handler/resource"
	authmiddleware "github.com/kkonst40/cloud-storage/backend/internal/api/middleware/auth"
	loggingmiddleware "github.com/kkonst40/cloud-storage/backend/internal/api/middleware/logging"
	validationmiddleware "github.com/kkonst40/cloud-storage/backend/internal/api/middleware/validation"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
	errs "github.com/kkonst40/cloud-storage/backend/internal/errors"
	"github.com/kkonst40/cloud-storage/backend/internal/service/session"
)

type Client struct {
	router *fiber.App
	port   string
}

func NewClient(
	cfg *config.Config,
	authHandler *authhandler.Handler,
	resourceHandler *resourcehandler.Handler,
	profileHandler *profilehandler.Handler,
	sessionService *session.Service,
) *Client {
	app := fiber.New(fiber.Config{BodyLimit: cfg.ApiFileUploadMaxSize * 1024 * 1024})
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(cfg.CORSAllowOrigins, ","),
		AllowMethods:     strings.Split(cfg.CORSAllowMethods, ","),
		AllowHeaders:     strings.Split(cfg.CORSAllowHeaders, ","),
		AllowCredentials: cfg.CORSAllowCredentials,
	}))

	app.Get("/api/user/me", authmiddleware.Auth(cfg, sessionService), profileHandler.Show)

	authGroup := app.Group("/api/auth")
	authGroup.Use(loggingmiddleware.Logging)
	authGroup.Post("/sign-in", validationmiddleware.CredentialsValidation, authHandler.Login)
	authGroup.Post("/sign-up", validationmiddleware.CredentialsValidation, authHandler.Register)
	authGroup.Post("/sign-out", authmiddleware.Auth(cfg, sessionService), authHandler.Logout)

	resourceGroup := app.Group("/api/resource")
	resourceGroup.Use(loggingmiddleware.Logging)
	resourceGroup.Use(authmiddleware.Auth(cfg, sessionService))
	resourceGroup.Get("/", resourceHandler.Show)
	resourceGroup.Post("/", resourceHandler.Store)
	resourceGroup.Delete("/", resourceHandler.Delete)
	resourceGroup.Get("/move", resourceHandler.Move)
	resourceGroup.Get("/download", resourceHandler.Download)
	resourceGroup.Get("/search", resourceHandler.Search)

	directoryGroup := app.Group("/api/directory")
	directoryGroup.Use(loggingmiddleware.Logging)
	directoryGroup.Use(authmiddleware.Auth(cfg, sessionService))
	directoryGroup.Get("/", resourceHandler.DirectoryShow)
	directoryGroup.Post("/", resourceHandler.DirectoryStore)

	return &Client{router: app, port: cfg.ApiPort}
}

func (cl *Client) Start() {
	go func() {
		err := cl.router.Listen(cl.port)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func (cl *Client) Shutdown() error {
	if err := cl.router.Shutdown(); err != nil {
		return errs.Wrap("api.Client", "Shutdown", err)
	}

	fmt.Println("API Server Shutdown")

	return nil
}
