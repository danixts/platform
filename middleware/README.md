# middleware

Fiber middlewares y helpers de identidad para microservicios.

## Setup

```go
app := fiber.New()
app.Use(middleware.RequestID())
app.Use(middleware.AccessLog())
app.Use(middleware.Recover())
app.Use(middleware.Tenant()) // extrae identity de headers X-*
```

### Opciones de Tenant

```go
app.Use(middleware.Tenant(middleware.Options{
    RequireValid:   true,
    RequireAccount: true,
    OnReject: func(c fiber.Ctx, reason string) error {
        return c.Status(401).JSON(fiber.Map{"error": reason})
    },
}))
```

## Leer identidad en handlers

Un campo:

```go
uid := middleware.GetUserUID(c)
org := middleware.GetOrgUID(c)
```

Múltiples campos (un solo lookup):

```go
id := middleware.GetIdentity(c)
uid := id.UserUID
org := id.OrgUID
role := id.Role
```

Todos los getters disponibles:

| Getter | Tipo | Header |
|--------|------|--------|
| `GetIsValid` | `bool` | `X-Is-Valid` |
| `GetAuthType` | `string` | `X-Auth-Type` |
| `GetUserUID` | `string` | `X-User-Uid` |
| `GetUsername` | `string` | `X-Username` |
| `GetEmail` | `string` | `X-Email` |
| `GetRole` | `string` | `X-Role` |
| `GetIsSuperAdmin` | `bool` | `X-Is-Super-Admin` |
| `GetIsInternal` | `bool` | `X-Is-Internal` |
| `GetUserStatus` | `string` | `X-User-Status` |
| `GetTimezone` | `string` | `X-Timezone` |
| `GetTimezoneOffset` | `string` | `X-Timezone-Offset` |
| `GetAccountUID` | `string` | `X-Account-Uid` |
| `GetOrgUID` | `string` | `X-Org-Uid` |
| `GetProductSlugs` | `[]string` | `X-Product-Slugs` |
| `GetRequestID` | `string` | `X-Request-Id` |

## Métodos de Identity

```go
id.IsAuthenticated()      // IsValid && UserUID != ""
id.HasProduct("billing")  // verifica slug en ProductSlugs
```

## Request ID standalone

Si no usás `Tenant`, el request ID está disponible por separado:

```go
app.Use(middleware.RequestID())

rid := middleware.RequestIDFromContext(c)
```
