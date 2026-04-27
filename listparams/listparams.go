// Package listparams parses pagination/sort/filter query params from a fiber
// request into a struct that handlers can map onto their own filter types.
//
// It deliberately does NOT depend on any service-specific domain.ListOpts —
// the handler is responsible for translating Result.Filters into typed fields
// (Status, CustomerUID, etc.). What this package guarantees is:
//
//   - page/limit are clamped to safe bounds (page >= 1, 1 <= limit <= 100).
//   - sort is whitelisted and emitted in canonical SQL form ("col DIR"),
//     safe to pass to gorm.Order(). Anything outside the whitelist falls
//     back to the default.
//   - filter keys are restricted to a per-handler whitelist; unknown query
//     params are dropped silently.
//
// Typical use:
//
//	var listSpec = listparams.Spec{
//	    AllowedFilters: []string{"status", "category_id"},
//	    AllowedSorts:   []string{"id", "created_at", "name"},
//	    DefaultSort:    "created_at DESC",
//	}
//
//	func (h *Handler) List(c fiber.Ctx) error {
//	    p := listparams.Parse(c, listSpec)
//	    items, total, err := h.uc.List(c.Context(), domain.ListOpts{
//	        Page: p.Page, Limit: p.Limit,
//	        OrderBy: p.Sort, Search: p.Search,
//	        Filters: p.Filters,
//	    })
//	    ...
//	}
package listparams

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Spec describes which query params a List handler accepts.
type Spec struct {
	// AllowedFilters lists query keys forwarded as exact-match filters into
	// Result.Filters. Convention: range filters use min_<field>/max_<field>
	// or date_from/date_to and are interpreted by the repo layer.
	AllowedFilters []string

	// AllowedSorts whitelists the column names accepted in `?sort=col:dir`.
	// Anything outside falls back to DefaultSort.
	AllowedSorts []string

	// DefaultSort is the SQL-form fallback (e.g. "created_at DESC"). Required.
	DefaultSort string
}

// Result is what handlers consume. Page/Limit are already clamped, Sort is
// already whitelisted, Filters only contains keys from Spec.AllowedFilters.
type Result struct {
	Page    int
	Limit   int
	Sort    string // canonical SQL form, safe for gorm.Order()
	Search  string
	Filters map[string]string
}

// Parse extracts page/limit/sort/q + whitelisted filters from the request.
func Parse(c fiber.Ctx, spec Spec) Result {
	page := fiber.Query[int](c, "page", 1)
	limit := fiber.Query[int](c, "limit", defaultLimit)
	page, limit = ClampPage(page, limit)

	r := Result{
		Page:    page,
		Limit:   limit,
		Sort:    ResolveSort(c, spec.AllowedSorts, spec.DefaultSort),
		Search:  strings.TrimSpace(c.Query("q")),
		Filters: map[string]string{},
	}
	for _, k := range spec.AllowedFilters {
		if v := strings.TrimSpace(c.Query(k)); v != "" {
			r.Filters[k] = v
		}
	}
	return r
}

// ResolveSort parses `?sort=<field>:<asc|desc>` against allowed and returns
// canonical SQL form ("field DIR"). Falls back to defaultSort if missing or
// not whitelisted.
func ResolveSort(c fiber.Ctx, allowed []string, defaultSort string) string {
	raw := strings.TrimSpace(c.Query("sort"))
	if raw == "" {
		return defaultSort
	}
	parts := strings.SplitN(raw, ":", 2)
	field := strings.ToLower(strings.TrimSpace(parts[0]))
	dir := "ASC"
	if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[1]), "desc") {
		dir = "DESC"
	}
	if !slices.Contains(allowed, field) {
		return defaultSort
	}
	return field + " " + dir
}

// ClampPage normalises page/limit (page >= 1, 1 <= limit <= 100, default 20).
func ClampPage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}
