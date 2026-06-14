// Package controllers implements the HTTP endpoints, middlewares, routing, and request handlers for the Veilo API server.
package controllers

import (
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

type RouteDeps struct {
	DomainSvc     services.DomainService
	AliasSvc      services.AliasService
	ForwardLogSvc services.ForwardLogService
	WebhookSvc    services.WebhookService
	WebhookSecret string
	APIKey        string
}

// APIKeyAuth provides simple bearer/api-key token auth middleware.
func APIKeyAuth(apiKey string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if apiKey == "" {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return SendError(c, fiber.StatusUnauthorized, "missing authorization header", nil)
		}

		token := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = authHeader[7:]
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(apiKey)) != 1 {
			return SendError(c, fiber.StatusUnauthorized, "invalid API key", nil)
		}

		return c.Next()
	}
}

// SetupRoutes registers all route groups by instantiating and registering controllers.
func SetupRoutes(app *fiber.App, deps RouteDeps) {
	// Apply API-key middleware to all /v1 API routes
	app.Use("/v1", APIKeyAuth(deps.APIKey))

	domainController := NewDomainController(deps.DomainSvc)
	domainController.RegisterRoutes(app)

	aliasController := NewAliasController(deps.AliasSvc, deps.ForwardLogSvc)
	aliasController.RegisterRoutes(app)

	forwardLogController := NewForwardLogController(deps.ForwardLogSvc, deps.AliasSvc)
	forwardLogController.RegisterRoutes(app)

	statsController := NewStatsController(deps.ForwardLogSvc)
	statsController.RegisterRoutes(app)

	webhookController := NewWebhookController(deps)
	webhookController.RegisterRoutes(app)
}
