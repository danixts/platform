// Package response provides the standard JSON envelope used by every
// service in the XMart Cloud platform. All API responses are wrapped in a
// Body struct so clients can rely on a single shape.
package response

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// Body is the envelope returned by every handler. Data is omitted from the
// JSON output when nil so error responses do not carry a null field.
type Body struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ValidationBody is the envelope used for 422 responses with a list of
// field-level validation errors.
type ValidationBody struct {
	Success bool     `json:"success"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

// OK writes a 200 response with the given data and message.
func OK(c fiber.Ctx, data any, message string) error {
	return c.Status(fiber.StatusOK).JSON(Body{
		Success: true, Code: fiber.StatusOK, Message: message, Data: data,
	})
}

// Created writes a 201 response.
func Created(c fiber.Ctx, data any, message string) error {
	return c.Status(fiber.StatusCreated).JSON(Body{
		Success: true, Code: fiber.StatusCreated, Message: message, Data: data,
	})
}

// NoContent writes a 204 response with an empty Body. Fiber strips the body
// automatically for 204 status.
func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// BadRequest writes a 400.
func BadRequest(c fiber.Ctx, message string) error {
	return fail(c, fiber.StatusBadRequest, message)
}

// Unauthorized writes a 401. An empty message defaults to "unauthorized".
func Unauthorized(c fiber.Ctx, message string) error {
	if message == "" {
		message = "unauthorized"
	}
	return fail(c, fiber.StatusUnauthorized, message)
}

// Forbidden writes a 403.
func Forbidden(c fiber.Ctx, message string) error {
	return fail(c, fiber.StatusForbidden, message)
}

// NotFound writes a 404. An empty message defaults to "not_found".
func NotFound(c fiber.Ctx, message string) error {
	if message == "" {
		message = "not_found"
	}
	return fail(c, fiber.StatusNotFound, message)
}

// Conflict writes a 409.
func Conflict(c fiber.Ctx, message string) error {
	return fail(c, fiber.StatusConflict, message)
}

// ValidationFail writes a 422 with a list of field-level errors.
func ValidationFail(c fiber.Ctx, errs []string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(ValidationBody{
		Success: false,
		Code:    fiber.StatusUnprocessableEntity,
		Message: "validation_failed",
		Errors:  errs,
	})
}

// Internal writes a 500 with a generic "internal_server_error" message.
// The underlying error should be logged by the caller — this package
// intentionally does not depend on a logger to stay framework-agnostic.
func Internal(c fiber.Ctx) error {
	return fail(c, fiber.StatusInternalServerError, "internal_server_error")
}

func fail(c fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(Body{
		Success: false, Code: code, Message: message, Data: nil,
	})
}

// Sentinel errors that services can return from their domain layer; the
// FromErr helper maps them to the right HTTP status.
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

// FromErr maps a domain error to the appropriate HTTP response. If the
// error does not match any sentinel, a 500 is returned.
func FromErr(c fiber.Ctx, err error) error {
	for sentinel, code := range errorStatus {
		if errors.Is(err, sentinel) {
			return fail(c, code, err.Error())
		}
	}
	return Internal(c)
}
