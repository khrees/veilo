package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

// IDomainController interface for domain controller
type IDomainController interface {
	RegisterRoutes(app *fiber.App)
	RegisterDomain(ctx fiber.Ctx) error
	ListDomains(ctx fiber.Ctx) error
	GetDomain(ctx fiber.Ctx) error
	RemoveDomain(ctx fiber.Ctx) error
}

type domainController struct {
	domainSvc services.IDomainService
}

// NewDomainController creates a new domain controller
func NewDomainController(domainSvc services.IDomainService) IDomainController {
	return &domainController{domainSvc: domainSvc}
}

func (c *domainController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/v1")

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
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	body.Domain = strings.TrimSpace(body.Domain)
	if body.Domain == "" {
		return fiber.NewError(fiber.StatusBadRequest, "domain name is required")
	}
	if !strings.Contains(body.Domain, ".") || strings.HasPrefix(body.Domain, ".") || strings.HasSuffix(body.Domain, ".") {
		return fiber.NewError(fiber.StatusBadRequest, "domain must be a valid domain name (e.g. example.com)")
	}

	if err := c.domainSvc.Register(body.Domain); err != nil {
		return err
	}
	return SendSuccess(ctx, fiber.StatusCreated, "Domain registered successfully", nil)
}

func (c *domainController) ListDomains(ctx fiber.Ctx) error {
	domains, err := c.domainSvc.FindAll()
	if err != nil {
		return err
	}
	return SendSuccess(ctx, fiber.StatusOK, "Domains retrieved successfully", domains)
}

func (c *domainController) GetDomain(ctx fiber.Ctx) error {
	d, err := c.domainSvc.FindByName(ctx.Params("domain"))
	if err != nil {
		return err
	}
	return SendSuccess(ctx, fiber.StatusOK, "Domain retrieved successfully", d)
}

func (c *domainController) RemoveDomain(ctx fiber.Ctx) error {
	if err := c.domainSvc.Remove(ctx.Params("domain")); err != nil {
		return err
	}
	return SendSuccess(ctx, fiber.StatusOK, "Domain removed successfully", nil)
}
