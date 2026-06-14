package controllers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/services"
)

const (
	defaultLimit  = 50
	maxLimit      = 100
	defaultOffset = 0
)

type forwardLogController struct {
	forwardLogSvc services.ForwardLogService
	aliasSvc      services.AliasService
}

// NewForwardLogController creates a new forward log controller
func NewForwardLogController(forwardLogSvc services.ForwardLogService, aliasSvc services.AliasService) *forwardLogController {
	return &forwardLogController{
		forwardLogSvc: forwardLogSvc,
		aliasSvc:      aliasSvc,
	}
}

func (c *forwardLogController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/v1")

	api.Get("/aliases/:aliasID/logs", c.GetForwardLogs)
}

func (c *forwardLogController) resolveAlias(idOrAddress string) (*models.Alias, error) {
	if c.aliasSvc == nil {
		parsedUUID, _ := uuid.Parse(idOrAddress)
		return &models.Alias{ID: parsedUUID, Address: idOrAddress}, nil
	}
	if strings.Contains(idOrAddress, "@") {
		return c.aliasSvc.FindByAddress(idOrAddress)
	}
	return c.aliasSvc.GetByID(idOrAddress)
}

func (c *forwardLogController) GetForwardLogs(ctx fiber.Ctx) error {
	aliasIDOrAddress := ctx.Params("aliasID")
	alias, err := c.resolveAlias(aliasIDOrAddress)
	if err != nil {
		return err
	}

	limit := defaultLimit
	offset := defaultOffset

	if l := ctx.Query("limit"); l != "" {
		limit = ClampInt(l, 1, maxLimit, defaultLimit)
	}
	if o := ctx.Query("offset"); o != "" {
		offset = ClampInt(o, 0, 1<<31-1, defaultOffset)
	}

	logs, err := c.forwardLogSvc.GetByAliasID(alias.ID.String(), limit, offset)
	if err != nil {
		return err
	}

	return SendSuccess(ctx, fiber.StatusOK, "Forward logs retrieved successfully", logs)
}

// ClampInt parses a string as an integer, returning fallback on error
// and clamping the result to [min, max].
func ClampInt(s string, min, max, fallback int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
