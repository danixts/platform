# db/postgres

Conexión Postgres con GORM y BaseRepo genérico.

## Conexión

```go
db, err := postgres.New(postgres.Config{
    DSN:      "postgres://user:pass@localhost:5432/mydb?sslmode=disable",
    LogLevel: postgres.LogWarn, // Silent | Error | Warn | Info
})
defer postgres.Close(db)
```

## BaseRepo

Embedding en repositorios propios:

```go
type UserRepo struct {
    postgres.BaseRepo[User]
}

func NewUserRepo(db *gorm.DB) *UserRepo {
    return &UserRepo{BaseRepo: postgres.NewBaseRepo[User](db)}
}
```

Métodos disponibles:

```go
repo.Create(ctx, &user)
repo.Save(ctx, &user)
repo.First(ctx, postgres.Where("uid = ?", uid))
repo.Find(ctx, postgres.Where("org_uid = ?", orgUID), postgres.Order("created_at DESC"))
repo.Count(ctx, postgres.Where("active = ?", true))
repo.UpdateWhere(ctx, map[string]any{"status": "inactive"}, postgres.Where("org_uid = ?", orgUID))
repo.DeleteWhere(ctx, postgres.Where("uid = ?", uid))
```

## Scopes disponibles

```go
postgres.Where(query, args...)
postgres.Order(order)
postgres.Limit(n)
postgres.Offset(n)
postgres.Preload(relation, args...)
```

Se pueden combinar:

```go
repo.Find(ctx,
    postgres.Where("org_uid = ?", orgUID),
    postgres.Where("active = ?", true),
    postgres.Order("created_at DESC"),
    postgres.Limit(20),
    postgres.Offset((page-1)*20),
)
```

## Acceso directo a GORM

Para queries complejas:

```go
repo.DB.WithContext(ctx).Raw("SELECT ...").Scan(&result)
```
