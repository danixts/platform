package response

import "github.com/gofiber/fiber/v3"

type Page[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func NewPage[T any](items []T, page, pageSize int, total int64) Page[T] {
	if pageSize < 1 {
		pageSize = 1
	}
	if page < 1 {
		page = 1
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if items == nil {
		items = []T{}
	}
	return Page[T]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

func OKPage[T any](c fiber.Ctx, page Page[T], message string) error {
	return c.Status(fiber.StatusOK).JSON(Body{
		Success: true, Code: fiber.StatusOK, Message: message, Data: page,
	})
}
