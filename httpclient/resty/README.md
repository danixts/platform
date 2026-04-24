# httpclient/resty

HTTP client fluent con reintentos, JSON automático y soporte de autenticación.

## Crear cliente

```go
svc := resty.NewService(resty.Config{
    BaseURL:       "https://api.example.com",
    Timeout:       15 * time.Second, // default: 15s
    RetryCount:    2,                // default: 2
    RetryWaitTime: 500 * time.Millisecond,
})
```

## Requests con genéricos (recomendado)

```go
// GET tipado
user, err := resty.Get[UserResp](svc.Call(ctx).Bearer(token), "/users/123")

// POST tipado
created, err := resty.Post[UserResp, CreateUserReq](svc.Call(ctx).Bearer(token), "/users", body)

// PUT / PATCH / DELETE
updated, err := resty.Put[UserResp, UpdateReq](svc.Call(ctx), "/users/123", body)
deleted, err := resty.Delete[UserResp](svc.Call(ctx), "/users/123")
```

## Builder de Call

```go
resp, err := svc.Call(ctx).
    Bearer(token).
    JSON().
    Body(payload).
    Into(&result).
    Strict().       // retorna error si status >= 400
    Post("/users")
```

### Métodos del builder

| Método | Descripción |
|--------|-------------|
| `Bearer(token)` | Authorization: Bearer |
| `Basic(user, pass)` | Authorization: Basic |
| `Authorization(value)` | header libre |
| `Header(key, values...)` | header custom |
| `JSON()` | Content-Type + Accept application/json |
| `Form()` | Content-Type form-urlencoded |
| `Body(payload)` | body del request |
| `Into(target)` | destino del unmarshal |
| `Strict()` | error en status >= 400 |

## Manejo de errores HTTP

```go
user, err := resty.Get[UserResp](svc.Call(ctx).Strict(), "/users/123")
if code, ok := resty.HTTPStatusCode(err); ok {
    // code == 404, 401, etc.
}
```
