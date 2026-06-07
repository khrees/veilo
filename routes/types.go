// Package routes registers all HTTP routes.
package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

// RouteDeps groups the services required to register the API routes.
type RouteDeps struct {
	DomainSvc     services.IDomainService
	AliasSvc      services.IAliasService
	ForwardLogSvc services.IForwardLogService
}

// IDomainController interface for domain controller
type IDomainController interface {
	RegisterRoutes(app *fiber.App)
	RegisterDomain(ctx fiber.Ctx) error
	ListDomains(ctx fiber.Ctx) error
	GetDomain(ctx fiber.Ctx) error
	RemoveDomain(ctx fiber.Ctx) error
}

// IAliasController interface for alias controller
type IAliasController interface {
	RegisterRoutes(app *fiber.App)
	CreateAlias(ctx fiber.Ctx) error
	ListAliases(ctx fiber.Ctx) error
	GetAlias(ctx fiber.Ctx) error
	UpdateAlias(ctx fiber.Ctx) error
	DeleteAlias(ctx fiber.Ctx) error
}

// IForwardLogController interface for forward log controller
type IForwardLogController interface {
	RegisterRoutes(app *fiber.App)
	GetForwardLogs(ctx fiber.Ctx) error
}

// IStatsController interface for stats controller
type IStatsController interface {
	RegisterRoutes(app *fiber.App)
	GetStats(ctx fiber.Ctx) error
}
