package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/khrees/veilo/services"
)

type statsController struct {
	forwardLogSvc services.ForwardLogService
}

// NewStatsController creates a new stats controller
func NewStatsController(forwardLogSvc services.ForwardLogService) *statsController {
	return &statsController{forwardLogSvc: forwardLogSvc}
}

func (c *statsController) RegisterRoutes(app *fiber.App) {
	api := app.Group("/v1")

	api.Get("/stats", c.GetStats)
}

func (c *statsController) GetStats(ctx fiber.Ctx) error {
	stats, err := c.forwardLogSvc.GetStats()
	if err != nil {
		return err
	}
	return SendSuccess(ctx, fiber.StatusOK, "Stats retrieved successfully", stats)
}
