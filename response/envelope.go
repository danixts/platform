package response

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/danixts/platform/logger"
)

type Body struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ValidationBody struct {
	Success bool     `json:"success"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

var (
	ErrNotFound     = errors.New("not_found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadRequest   = errors.New("bad_request")
)

var errorStatus = map[error]int{
	ErrNotFound:     fiber.StatusNotFound,
	ErrConflict:     fiber.StatusConflict,
	ErrUnauthorized: fiber.StatusUnauthorized,
	ErrForbidden:    fiber.StatusForbidden,
	ErrBadRequest:   fiber.StatusBadRequest,
}

func OK(c fiber.Ctx, data any, message string) error {
	return c.Status(fiber.StatusOK).JSON(Body{
		Success: true, Code: fiber.StatusOK, Message: message, Data: data,
	})
}

func Created(c fiber.Ctx, data any, message string) error {
	return c.Status(fiber.StatusCreated).JSON(Body{
		Success: true, Code: fiber.StatusCreated, Message: message, Data: data,
	})
}

func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func BadRequest(c fiber.Ctx, message string) error {
	return fail(c, fiber.StatusBadRequest, message)
}

func Unauthorized(c fiber.Ctx, message string) error {
	if message == "" {
		message = "unauthorized"
	}
	return fail(c, fiber.StatusUnauthorized, message)
}

func Forbidden(c fiber.Ctx, message string) error {
	return fail(c, fiber.StatusForbidden, message)
}

func NotFound(c fiber.Ctx, message string) error {
	if message == "" {
		message = "not_found"
	}
	return fail(c, fiber.StatusNotFound, message)
}

func Conflict(c fiber.Ctx, message string) error {
	return fail(c, fiber.StatusConflict, message)
}

func ValidationFail(c fiber.Ctx, errs []string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(ValidationBody{
		Success: false,
		Code:    fiber.StatusUnprocessableEntity,
		Message: "validation_failed",
		Errors:  errs,
	})
}

func Internal(c fiber.Ctx, errs ...error) error {
	if len(errs) > 0 && errs[0] != nil {
		logger.Error().
			Err(errs[0]).
			Str("path", c.Path()).
			Str("method", c.Method()).
			Msg("handler internal error")
	}
	return fail(c, fiber.StatusInternalServerError, "internal_server_error")
}

func FromErr(c fiber.Ctx, err error) error {
	for sentinel, code := range errorStatus {
		if errors.Is(err, sentinel) {
			return fail(c, code, err.Error())
		}
	}
	return Internal(c, err)
}

func fail(c fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(Body{
		Success: false, Code: code, Message: message, Data: nil,
	})
}
