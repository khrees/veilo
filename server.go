package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/google/uuid"
	"github.com/khrees/cloakee/routes"
)

type server struct {
	app             *fiber.App
	port            string
	shutdownTimeout time.Duration
}

func NewServer(port string, deps routes.RouteDeps) *server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	registerMiddleware(app)
	registerRoutes(app, deps)

	return &server{
		app:             app,
		port:            port,
		shutdownTimeout: 5 * time.Second,
	}
}

func (s *server) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := s.app.Listen(":" + s.port); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.app.ShutdownWithContext(shutdownCtx); err != nil {
		if !errors.Is(err, context.Canceled) {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}
	}

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func registerMiddleware(app *fiber.App) {
	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.NewString()
		},
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))
}

func registerRoutes(app *fiber.App, deps routes.RouteDeps) {
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	routes.SetupRoutes(app, deps)
}
