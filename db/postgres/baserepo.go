package postgres

import (
	"context"

	"gorm.io/gorm"
)

// Scope is a composable GORM query builder fragment. Helpers like Where,
// Order and Preload return Scope values that BaseRepo applies in order
// to build the final query. Services can define their own Scope
// constructors for domain-specific filters.
type Scope func(*gorm.DB) *gorm.DB

// Where returns a Scope that adds a WHERE clause.
func Where(query any, args ...any) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Where(query, args...) }
}

// Order returns a Scope that adds an ORDER BY clause.
func Order(order string) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Order(order) }
}

// Limit returns a Scope that caps the number of rows returned.
func Limit(n int) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Limit(n) }
}

// Offset returns a Scope that skips the first n rows.
func Offset(n int) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Offset(n) }
}

// Preload returns a Scope that eager-loads the named association.
func Preload(relation string, args ...any) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Preload(relation, args...) }
}

// BaseRepo is a generic repository helper that wraps a *gorm.DB for a
// concrete entity type T. It provides Create/Save/First/Find/Count/
// UpdateWhere/DeleteWhere built on top of the Scope combinator.
//
// It is intended as a starting point for feature-specific repositories
// that want the boilerplate handled once. When the abstraction gets in
// the way, callers reach into BaseRepo.DB directly.
type BaseRepo[T any] struct {
	DB *gorm.DB
}

// NewBaseRepo builds a BaseRepo[T] bound to db.
func NewBaseRepo[T any](db *gorm.DB) BaseRepo[T] {
	return BaseRepo[T]{DB: db}
}

// Create inserts a new entity.
func (r *BaseRepo[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

// Save upserts an entity (insert or update by primary key).
func (r *BaseRepo[T]) Save(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Save(entity).Error
}

// First returns the first row matching the given scopes, or an error
// from gorm (including gorm.ErrRecordNotFound).
func (r *BaseRepo[T]) First(ctx context.Context, scopes ...Scope) (*T, error) {
	var entity T
	q := r.DB.WithContext(ctx)
	for _, s := range scopes {
		q = s(q)
	}
	if err := q.First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Find returns every row matching the given scopes. An empty result is
// returned as a nil slice, not an error.
func (r *BaseRepo[T]) Find(ctx context.Context, scopes ...Scope) ([]*T, error) {
	var entities []*T
	q := r.DB.WithContext(ctx)
	for _, s := range scopes {
		q = s(q)
	}
	if err := q.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Count returns the number of rows matching the given scopes.
func (r *BaseRepo[T]) Count(ctx context.Context, scopes ...Scope) (int64, error) {
	var count int64
	q := r.DB.WithContext(ctx).Model(new(T))
	for _, s := range scopes {
		q = s(q)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateWhere applies a partial update to every row matching the scopes.
func (r *BaseRepo[T]) UpdateWhere(ctx context.Context, values any, scopes ...Scope) error {
	q := r.DB.WithContext(ctx).Model(new(T))
	for _, s := range scopes {
		q = s(q)
	}
	return q.Updates(values).Error
}

// DeleteWhere deletes every row matching the scopes.
func (r *BaseRepo[T]) DeleteWhere(ctx context.Context, scopes ...Scope) error {
	q := r.DB.WithContext(ctx)
	for _, s := range scopes {
		q = s(q)
	}
	return q.Delete(new(T)).Error
}
