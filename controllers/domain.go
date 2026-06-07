package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

type domainController struct {
	domainSvc services.IDomainService
}

// NewDomainController creates a new domain controller
func NewDomainController(domainSvc services.IDomainService) IDomainController {
	return &domainController{domainSvc: domainSvc}
}

func (c *domainController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/domains", c.RegisterDomain)
	api.Get("/domains", c.ListDomains)
	api.Get("/domains/:domain", c.GetDomain)
	api.Delete("/domains/:domain", c.RemoveDomain)
}

func (c *domainController) RegisterDomain(ctx fiber.Ctx) error {
	var body struct {
		Domain string `json:"domain"`
	}
	if err := ctx.Bind().Body(&body); err != nil {
		return SendError(ctx, fiber.StatusBadRequest, "Invalid request body", err)
	}
	if err := c.domainSvc.Register(body.Domain); err != nil {
		return SendError(ctx, fiber.StatusInternalServerError, "Failed to register domain", err)
	}
	return SendSuccess(ctx, fiber.StatusCreated, "Domain registered successfully", nil)
}

func (c *domainController) ListDomains(ctx fiber.Ctx) error {
	domains, err := c.domainSvc.FindAll()
	if err != nil {
		return SendError(ctx, fiber.StatusInternalServerError, "Failed to list domains", err)
	}
	return SendSuccess(ctx, fiber.StatusOK, "Domains retrieved successfully", domains)
}

func (c *domainController) GetDomain(ctx fiber.Ctx) error {
	domain := ctx.Params("domain")
	d, err := c.domainSvc.FindByName(domain)
	if err != nil {
		return SendError(ctx, fiber.StatusNotFound, "Domain not found", err)
	}
	return SendSuccess(ctx, fiber.StatusOK, "Domain retrieved successfully", d)
}

func (c *domainController) RemoveDomain(ctx fiber.Ctx) error {
	if err := c.domainSvc.Remove(ctx.Params("domain")); err != nil {
		return SendError(ctx, fiber.StatusInternalServerError, "Failed to remove domain", err)
	}
	return SendSuccess(ctx, fiber.StatusOK, "Domain removed successfully", nil)
}
