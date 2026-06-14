package controllers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/services"
)

// IAliasController interface for alias controller
type IAliasController interface {
	RegisterRoutes(app *fiber.App)
	CreateAlias(ctx fiber.Ctx) error
	ListAliases(ctx fiber.Ctx) error
	GetAlias(ctx fiber.Ctx) error
	UpdateAlias(ctx fiber.Ctx) error
	DeleteAlias(ctx fiber.Ctx) error
}

type aliasController struct {
	aliasSvc services.IAliasService
}

// NewAliasController creates a new alias controller
func NewAliasController(aliasSvc services.IAliasService) IAliasController {
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
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	body.Address = strings.TrimSpace(body.Address)
	body.Slug = strings.TrimSpace(body.Slug)
	body.Domain = strings.TrimSpace(body.Domain)
	body.RealEmail = strings.TrimSpace(body.RealEmail)

	// If slug is empty but address is provided, extract slug from address
	if body.Slug == "" && body.Address != "" {
		parts := strings.Split(body.Address, "@")
		if len(parts) > 0 {
			body.Slug = parts[0]
		}
	}

	// If slug is still empty, generate a creative one
	if body.Slug == "" {
		body.Slug = generateCreativeSlug()
	}

	// If address is empty, construct it using slug and domain
	if body.Address == "" {
		if body.Domain == "" {
			return fiber.NewError(fiber.StatusBadRequest, "domain is required to generate address when address is empty")
		}
		body.Address = fmt.Sprintf("%s@%s", body.Slug, body.Domain)
	}

	if body.Address == "" || body.Slug == "" || body.Domain == "" || body.RealEmail == "" {
		return fiber.NewError(fiber.StatusBadRequest, "address, slug, domain, and real_email are required")
	}
	if _, err := mail.ParseAddress(body.RealEmail); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "real_email must be a valid email address")
	}
	if _, err := mail.ParseAddress(body.Address); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "address must be a valid email address")
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
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
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

var adjectives = []string{
	"glowing", "radiant", "whispering", "silent", "frosty", "golden",
	"silver", "crimson", "azure", "mystic", "shadowy", "stellar",
	"cosmic", "wild", "gentle", "bouncy", "jolly", "merry", "speedy",
	"vibrant", "serene", "dusk", "dawn", "misty", "stormy", "cloudy",
}

var nouns = []string{
	"umbrella", "forest", "sunset", "river", "mountain", "ocean",
	"breeze", "galaxy", "comet", "nebula", "meadow", "canyon",
	"beacon", "harbor", "castle", "fortress", "glade", "oasis",
	"pioneer", "valley", "summit", "island", "desert", "tundra",
}

func generateCreativeSlug() string {
	adjIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	nounIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	num, _ := rand.Int(rand.Reader, big.NewInt(1000)) // 0 to 999

	slug := fmt.Sprintf("%s-%s-%03d", adjectives[adjIdx.Int64()], nouns[nounIdx.Int64()], num.Int64())
	if len(slug) > 25 {
		return slug[:25]
	}
	return slug
}
