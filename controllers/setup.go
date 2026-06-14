package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

type RouteDeps struct {
	DomainSvc     services.IDomainService
	AliasSvc      services.IAliasService
	ForwardLogSvc services.IForwardLogService
	WebhookSvc    services.IWebhookService
	WebhookSecret string
}

// SetupRoutes registers all route groups by instantiating and registering controllers.
func SetupRoutes(app *fiber.App, deps RouteDeps) {
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
