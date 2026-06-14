package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

type RouteDeps struct {
	DomainSvc     services.IDomainService
	AliasSvc      services.IAliasService
	ForwardLogSvc services.IForwardLogService
	WebhookSvc    services.IWebhookService
	WebhookSecret string
	APIKey        string
}

// ApiKeyAuth provides simple bearer/api-key token auth middleware.
func ApiKeyAuth(apiKey string) fiber.Handler {
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

		if token != apiKey {
			return SendError(c, fiber.StatusUnauthorized, "invalid API key", nil)
		}

		return c.Next()
	}
}

// SetupRoutes registers all route groups by instantiating and registering controllers.
func SetupRoutes(app *fiber.App, deps RouteDeps) {
	// Apply API-key middleware to all /v1 API routes
	app.Use("/v1", ApiKeyAuth(deps.APIKey))

	domainController := NewDomainController(deps.DomainSvc)
	domainController.RegisterRoutes(app)

	aliasController := NewAliasController(deps.AliasSvc)
	aliasController.RegisterRoutes(app)

	forwardLogController := NewForwardLogController(deps.ForwardLogSvc)
	forwardLogController.RegisterRoutes(app)

	statsController := NewStatsController(deps.ForwardLogSvc)
	statsController.RegisterRoutes(app)

	webhookController := NewWebhookController(deps)
	webhookController.RegisterRoutes(app)
}
