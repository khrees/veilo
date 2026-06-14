package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "embed"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/google/uuid"
	"github.com/khrees/veilo/controllers"
	"gorm.io/gorm"
)

//go:embed install.sh
var installScript string

// ServerConfig groups configuration values needed by the HTTP server.
type ServerConfig struct {
	Port        string
	CORSOrigins []string
	RateLimit   int
}

type server struct {
	app             *fiber.App
	port            string
	shutdownTimeout time.Duration
}

func NewServer(cfg ServerConfig, deps controllers.RouteDeps) *server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		BodyLimit:    1 << 20,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "an internal server error occurred"

			// Check if it's a record not found error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return controllers.SendError(ctx, fiber.StatusNotFound, "resource not found", nil)
			}

			// Check for unique constraint / duplicate key errors
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") {
				return controllers.SendError(ctx, fiber.StatusConflict, "resource already exists", nil)
			}

			// Retrieve the custom status code if it's a *fiber.Error (e.g. validation errors)
			var e *fiber.Error
			if errors.As(err, &e) {
				return controllers.SendError(ctx, e.Code, e.Message, nil)
			}

			// Log the actual raw database or system error internally
			log.Errorf("Internal system error: %v", err)

			// Return a generic internal error message to the client
			return controllers.SendError(ctx, code, message, nil)
		},
	})
	registerMiddleware(app, cfg)
	registerRoutes(app, deps)

	return &server{
		app:             app,
		port:            cfg.Port,
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

func registerMiddleware(app *fiber.App, cfg ServerConfig) {
	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.NewString()
		},
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))
	app.Use(recover.New())
	app.Use(func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return c.Next()
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))
	app.Use(limiter.New(limiter.Config{
		Max:               cfg.RateLimit,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
}

func registerRoutes(app *fiber.App, deps controllers.RouteDeps) {
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Get("/install", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; charset=utf-8")
		return c.SendString(installScript)
	})

	app.Get("/install.sh", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; charset=utf-8")
		return c.SendString(installScript)
	})

	controllers.SetupRoutes(app, deps)
}
