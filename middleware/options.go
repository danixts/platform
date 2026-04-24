package middleware

import "github.com/gofiber/fiber/v3"

type Options struct {
	RequireValid   bool
	RequireAccount bool
	OnReject       func(c fiber.Ctx, reason string) error
}

func DefaultOptions() Options {
	return Options{
		RequireValid:   true,
		RequireAccount: true,
	}
}
