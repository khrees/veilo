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

// --------------------
// Domain Controller
// --------------------

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
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := c.domainSvc.Register(body.Domain); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusCreated)
}

func (c *domainController) ListDomains(ctx fiber.Ctx) error {
	domains, err := c.domainSvc.FindAll()
	if err != nil {
		return err
	}
	return ctx.JSON(domains)
}

func (c *domainController) GetDomain(ctx fiber.Ctx) error {
	domain := ctx.Params("domain")
	d, err := c.domainSvc.FindByDomain(domain)
	if err != nil {
		return err
	}
	return ctx.JSON(d)
}

func (c *domainController) RemoveDomain(ctx fiber.Ctx) error {
	if err := c.domainSvc.Remove(ctx.Params("domain")); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// --------------------
// Alias Controller
// --------------------

type aliasController struct {
	aliasSvc services.IAliasService
}

// NewAliasController creates a new alias controller
func NewAliasController(aliasSvc services.IAliasService) IAliasController {
	return &aliasController{aliasSvc: aliasSvc}
}

func (c *aliasController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/aliases", c.CreateAlias)
	api.Get("/aliases", c.ListAliases)
	api.Get("/aliases/:id", c.GetAlias)
	api.Put("/aliases/:id", c.UpdateAlias)
	api.Delete("/aliases/:id", c.DeleteAlias)
}

func (c *aliasController) CreateAlias(ctx fiber.Ctx) error {
	var body struct {
		Address   string  `json:"address"`
		Slug      string  `json:"slug"`
		Domain    string  `json:"domain"`
		RealEmail string  `json:"real_email"`
		Label     *string `json:"label,omitempty"`
		Enabled   *bool   `json:"enabled,omitempty"`
	}
	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	alias, err := c.aliasSvc.Create(services.AliasCreateInput{
		Address:   body.Address,
		Slug:      body.Slug,
		Domain:    body.Domain,
		RealEmail: body.RealEmail,
		Label:     body.Label,
		Enabled:   enabled,
	})
	if err != nil {
		return err
	}

	return ctx.JSON(alias)
}

func (c *aliasController) ListAliases(ctx fiber.Ctx) error {
	aliases, err := c.aliasSvc.GetAll()
	if err != nil {
		return err
	}
	return ctx.JSON(aliases)
}

func (c *aliasController) GetAlias(ctx fiber.Ctx) error {
	alias, err := c.aliasSvc.GetByID(ctx.Params("id"))
	if err != nil {
		return err
	}
	return ctx.JSON(alias)
}

func (c *aliasController) UpdateAlias(ctx fiber.Ctx) error {
	id := ctx.Params("id")

	var body struct {
		Address      *string `json:"address,omitempty"`
		RealEmail    *string `json:"real_email,omitempty"`
		Label        *string `json:"label,omitempty"`
		Enabled      *bool   `json:"enabled,omitempty"`
		ForwardCount *int    `json:"forward_count,omitempty"`
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updates := make(map[string]any)
	if body.Address != nil {
		updates["address"] = *body.Address
	}
	if body.RealEmail != nil {
		updates["real_email"] = *body.RealEmail
	}
	if body.Label != nil {
		updates["label"] = *body.Label
	}
	if body.Enabled != nil {
		updates["enabled"] = *body.Enabled
	}
	if body.ForwardCount != nil {
		updates["forward_count"] = *body.ForwardCount
	}

	if err := c.aliasSvc.Update(id, updates); err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *aliasController) DeleteAlias(ctx fiber.Ctx) error {
	if err := c.aliasSvc.Delete(ctx.Params("id")); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// --------------------
// Forward Log Controller
// --------------------

type forwardLogController struct {
	forwardLogSvc services.IForwardLogService
}

// NewForwardLogController creates a new forward log controller
func NewForwardLogController(forwardLogSvc services.IForwardLogService) IForwardLogController {
	return &forwardLogController{forwardLogSvc: forwardLogSvc}
}

func (c *forwardLogController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/forward-logs/:aliasID", c.GetForwardLogs)
}

func (c *forwardLogController) GetForwardLogs(ctx fiber.Ctx) error {
	aliasID := ctx.Params("aliasID")

	limit := 50
	offset := 0

	if l := ctx.Query("limit"); l != "" {
		limit = ParseInt(l)
	}
	if o := ctx.Query("offset"); o != "" {
		offset = ParseInt(o)
	}

	logs, err := c.forwardLogSvc.GetByAliasID(aliasID, limit, offset)
	if err != nil {
		return err
	}

	return ctx.JSON(logs)
}

// ParseInt converts a string to an integer, stopping at non-digit characters.
// It returns 0 if the string doesn't start with a digit.
func ParseInt(value string) int {
	var result int
	for _, c := range value {
		if c < '0' || c > '9' {
			break
		}
		result = result*10 + int(c-'0')
	}
	return result
}

// --------------------
// Stats Controller
// --------------------

type statsController struct {
	forwardLogSvc services.IForwardLogService
}

// NewStatsController creates a new stats controller
func NewStatsController(forwardLogSvc services.IForwardLogService) IStatsController {
	return &statsController{forwardLogSvc: forwardLogSvc}
}

func (c *statsController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/stats", c.GetStats)
}

func (c *statsController) GetStats(ctx fiber.Ctx) error {
	stats, err := c.forwardLogSvc.GetStats()
	if err != nil {
		return err
	}
	return ctx.JSON(stats)
}

// --------------------
// SetupRoutes - Register all controllers
// --------------------

// SetupRoutes registers all route groups by instantiating and registering controllers
func SetupRoutes(app *fiber.App, deps RouteDeps) {
	// Create and register domain controller
	domainController := NewDomainController(deps.DomainSvc)
	domainController.RegisterRoutes(app)

	// Create and register alias controller
	aliasController := NewAliasController(deps.AliasSvc)
	aliasController.RegisterRoutes(app)

	// Create and register forward log controller
	forwardLogController := NewForwardLogController(deps.ForwardLogSvc)
	forwardLogController.RegisterRoutes(app)

	// Create and register stats controller
	statsController := NewStatsController(deps.ForwardLogSvc)
	statsController.RegisterRoutes(app)
}
