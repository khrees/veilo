// Package routes registers all HTTP routes.
package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/khrees/cloakee/repositories"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/domains", func(c fiber.Ctx) error {
		return c.JSON([]repositories.Domain{
			{ID: "1", Name: "example.com"},
		})
	})
}
