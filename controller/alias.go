package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

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
