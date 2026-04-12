package postgres

import (
	"context"

	"gorm.io/gorm"
)

type Scope func(*gorm.DB) *gorm.DB

func Where(query any, args ...any) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Where(query, args...) }
}

func Order(order string) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Order(order) }
}

func Limit(n int) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Limit(n) }
}

func Offset(n int) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Offset(n) }
}

func Preload(relation string, args ...any) Scope {
	return func(db *gorm.DB) *gorm.DB { return db.Preload(relation, args...) }
}

type BaseRepo[T any] struct {
	DB *gorm.DB
}

func NewBaseRepo[T any](db *gorm.DB) BaseRepo[T] {
	return BaseRepo[T]{DB: db}
}

func (r *BaseRepo[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

func (r *BaseRepo[T]) Save(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Save(entity).Error
}

func (r *BaseRepo[T]) First(ctx context.Context, scopes ...Scope) (*T, error) {
	var entity T
	query := r.DB.WithContext(ctx)
	for _, scope := range scopes {
		query = scope(query)
	}
	if err := query.First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepo[T]) Find(ctx context.Context, scopes ...Scope) ([]*T, error) {
	var entities []*T
	query := r.DB.WithContext(ctx)
	for _, scope := range scopes {
		query = scope(query)
	}
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *BaseRepo[T]) Count(ctx context.Context, scopes ...Scope) (int64, error) {
	var count int64
	query := r.DB.WithContext(ctx).Model(new(T))
	for _, scope := range scopes {
		query = scope(query)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *BaseRepo[T]) UpdateWhere(ctx context.Context, values any, scopes ...Scope) error {
	query := r.DB.WithContext(ctx).Model(new(T))
	for _, scope := range scopes {
		query = scope(query)
	}
	return query.Updates(values).Error
}

func (r *BaseRepo[T]) DeleteWhere(ctx context.Context, scopes ...Scope) error {
	query := r.DB.WithContext(ctx)
	for _, scope := range scopes {
		query = scope(query)
	}
	return query.Delete(new(T)).Error
}
