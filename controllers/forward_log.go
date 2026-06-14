package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

// IForwardLogController interface for forward log controller
type IForwardLogController interface {
	RegisterRoutes(app *fiber.App)
	GetForwardLogs(ctx fiber.Ctx) error
}

type forwardLogController struct {
	forwardLogSvc services.IForwardLogService
}

// NewForwardLogController creates a new forward log controller
func NewForwardLogController(forwardLogSvc services.IForwardLogService) IForwardLogController {
	return &forwardLogController{forwardLogSvc: forwardLogSvc}
}

func (c *forwardLogController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/v1")

	api.Get("/forward-logs/:aliasID", c.GetForwardLogs)
	api.Get("/aliases/:aliasID/logs", c.GetForwardLogs)
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

	return SendSuccess(ctx, fiber.StatusOK, "Forward logs retrieved successfully", logs)
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
