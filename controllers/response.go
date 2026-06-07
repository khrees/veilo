package controllers

import (
	"github.com/gofiber/fiber/v3"
)

// Response represents the standard API response structure.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SendSuccess sends a successful JSON response with a given status code.
func SendSuccess(ctx fiber.Ctx, status int, message string, data interface{}) error {
	return ctx.Status(status).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendError sends a failure JSON response with a given status code.
func SendError(ctx fiber.Ctx, status int, message string, err interface{}) error {
	return ctx.Status(status).JSON(Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}
