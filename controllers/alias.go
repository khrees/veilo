package controllers

import (
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/services"
)

type aliasController struct {
	aliasSvc services.AliasService
}

// NewAliasController creates a new alias controller
func NewAliasController(aliasSvc services.AliasService) *aliasController {
	return &aliasController{aliasSvc: aliasSvc}
}

func (c *aliasController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/v1")

	api.Post("/aliases", c.CreateAlias)
	api.Get("/aliases", c.ListAliases)
	api.Get("/aliases/:id", c.GetAlias)
	api.Put("/aliases/:id", c.UpdateAlias)
	api.Delete("/aliases/:id", c.DeleteAlias)
}

func (c *aliasController) CreateAlias(ctx fiber.Ctx) error {
	var body struct {
		Address     string  `json:"address"`
		Slug        string  `json:"slug"`
		Domain      string  `json:"domain"`
		RealEmail   string  `json:"real_email"`
		DisplayName *string `json:"display_name,omitempty"`
		Label       *string `json:"label,omitempty"`
		Enabled     *bool   `json:"enabled,omitempty"`
	}
	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	body.Address = strings.TrimSpace(body.Address)
	body.Slug = strings.TrimSpace(body.Slug)
	body.Domain = strings.ToLower(strings.TrimSpace(body.Domain))
	body.RealEmail = strings.TrimSpace(body.RealEmail)

	if body.Domain == "" || body.RealEmail == "" {
		return fiber.NewError(fiber.StatusBadRequest, "domain and real_email are required")
	}
	if _, err := mail.ParseAddress(body.RealEmail); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "real_email must be a valid email address")
	}
	if body.Address != "" {
		if _, err := mail.ParseAddress(body.Address); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "address must be a valid email address")
		}
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	alias, err := c.aliasSvc.Create(services.AliasCreateInput{
		Address:     body.Address,
		Slug:        body.Slug,
		Domain:      body.Domain,
		RealEmail:   body.RealEmail,
		DisplayName: body.DisplayName,
		Label:       body.Label,
		Enabled:     enabled,
	})
	if err != nil {
		return err
	}

	return SendSuccess(ctx, fiber.StatusCreated, "Alias created successfully", alias)
}

func (c *aliasController) ListAliases(ctx fiber.Ctx) error {
	var filter models.AliasFilter

	if enabledStr := ctx.Query("enabled"); enabledStr != "" {
		switch enabledStr {
		case "true":
			val := true
			filter.Enabled = &val
		case "false":
			val := false
			filter.Enabled = &val
		}
	}

	if domainStr := ctx.Query("domain"); domainStr != "" {
		filter.Domain = &domainStr
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		val := ClampInt(limitStr, 1, maxLimit, defaultLimit)
		filter.Limit = &val
	}

	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		val := ClampInt(offsetStr, 0, 1<<31-1, defaultOffset)
		filter.Offset = &val
	}

	aliases, err := c.aliasSvc.GetAll(filter)
	if err != nil {
		return err
	}
	return SendSuccess(ctx, fiber.StatusOK, "Aliases retrieved successfully", aliases)
}

func (c *aliasController) GetAlias(ctx fiber.Ctx) error {
	alias, err := c.aliasSvc.GetByID(ctx.Params("id"))
	if err != nil {
		return err
	}
	return SendSuccess(ctx, fiber.StatusOK, "Alias retrieved successfully", alias)
}

func (c *aliasController) UpdateAlias(ctx fiber.Ctx) error {
	id := ctx.Params("id")

	var body struct {
		Address     *string `json:"address,omitempty"`
		RealEmail   *string `json:"real_email,omitempty"`
		DisplayName *string `json:"display_name,omitempty"`
		Label       *string `json:"label,omitempty"`
		Enabled     *bool   `json:"enabled,omitempty"`
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	updates := make(map[string]any)
	if body.Address != nil {
		addr := strings.TrimSpace(*body.Address)
		if _, err := mail.ParseAddress(addr); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "address must be a valid email address")
		}
		updates["address"] = addr
	}
	if body.RealEmail != nil {
		email := strings.TrimSpace(*body.RealEmail)
		if _, err := mail.ParseAddress(email); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "real_email must be a valid email address")
		}
		updates["real_email"] = email
	}
	if body.DisplayName != nil {
		updates["display_name"] = *body.DisplayName
	}
	if body.Label != nil {
		updates["label"] = *body.Label
	}
	if body.Enabled != nil {
		updates["enabled"] = *body.Enabled
	}

	if err := c.aliasSvc.Update(id, updates); err != nil {
		return err
	}

	return SendSuccess(ctx, fiber.StatusOK, "Alias updated successfully", nil)
}

func (c *aliasController) DeleteAlias(ctx fiber.Ctx) error {
	if err := c.aliasSvc.Delete(ctx.Params("id")); err != nil {
		return err
	}
	return SendSuccess(ctx, fiber.StatusOK, "Alias deleted successfully", nil)
}
