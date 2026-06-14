package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

const (
	defaultLimit  = 50
	maxLimit      = 100
	defaultOffset = 0
)

type forwardLogController struct {
	forwardLogSvc services.ForwardLogService
}

// NewForwardLogController creates a new forward log controller
func NewForwardLogController(forwardLogSvc services.ForwardLogService) *forwardLogController {
	return &forwardLogController{forwardLogSvc: forwardLogSvc}
}

func (c *forwardLogController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/v1")

	api.Get("/aliases/:aliasID/logs", c.GetForwardLogs)
}

func (c *forwardLogController) GetForwardLogs(ctx fiber.Ctx) error {
	aliasID := ctx.Params("aliasID")

	limit := defaultLimit
	offset := defaultOffset

	if l := ctx.Query("limit"); l != "" {
		limit = ClampInt(l, 1, maxLimit, defaultLimit)
	}
	if o := ctx.Query("offset"); o != "" {
		offset = ClampInt(o, 0, 1<<31-1, defaultOffset)
	}

	logs, err := c.forwardLogSvc.GetByAliasID(aliasID, limit, offset)
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
