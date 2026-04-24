# response

Helpers de respuesta HTTP para handlers Fiber. Envelope uniforme en todos los ms.

## Respuestas exitosas

```go
response.OK(c, data, "ok")
response.Created(c, data, "created")
response.NoContent(c)
```

## Respuesta paginada

```go
page := response.NewPage(items, pageNum, pageSize, total)
response.OKPage(c, page, "ok")
```

`Page[T]` devuelve:

```json
{
  "success": true,
  "code": 200,
  "message": "ok",
  "data": {
    "items": [...],
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## Errores HTTP

```go
response.BadRequest(c, "invalid_payload")
response.Unauthorized(c, "")        // mensaje default: "unauthorized"
response.Forbidden(c, "no_access")
response.NotFound(c, "")            // mensaje default: "not_found"
response.Conflict(c, "already_exists")
response.ValidationFail(c, []string{"name is required", "email is invalid"})
response.Internal(c, err)           // loggea el error, devuelve 500 genérico
```

## Mapeo automático desde errores de dominio

Usar en el handler cuando el use case devuelve errores sentinel:

```go
user, err := uc.GetUser(ctx, id)
if err != nil {
    return response.FromErr(c, err)
}
return response.OK(c, user, "ok")
```

Sentinels soportados:

| Error | Status |
|-------|--------|
| `response.ErrNotFound` | 404 |
| `response.ErrConflict` | 409 |
| `response.ErrUnauthorized` | 401 |
| `response.ErrForbidden` | 403 |
| `response.ErrBadRequest` | 400 |

Cualquier otro error → 500 con log automático.
