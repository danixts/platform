# platform

Opinionated Go SDK for building HTTP services with Fiber v3. Ships thin
wrappers over the usual suspects — Postgres (GORM), Redis, NATS, S3,
HTTP client — plus a request-scoped identity middleware, a standard
response envelope, structured logging and a handful of utilities.

## Design goals

- **Thin wrappers, not frameworks.** Every client exposes `New(Config)`
  and a small surface. When the wrapper is not enough, `Raw()` returns
  the underlying library client so you can use it directly.
- **Context-first.** All I/O takes `context.Context` as its first
  parameter.
- **No hidden panics.** The only exception is `MustFromContext`, which
  panics when the middleware chain is misconfigured (a programming
  error, not a runtime condition).
- **Config via plain structs.** No global state, no environment parsing
  inside the SDK — the consumer owns configuration loading.
- **Semantic versioning.** Pre-1.0, minor bumps may break API; from
  v1.0 onwards, breaking changes only in major versions.

## Packages

| Package | Purpose |
|---|---|
| `middleware` | Fiber v3 handlers for identity (`Tenant`), request id, access log and panic recovery. |
| `response` | Standard `Body` envelope, generic `Page[T]`, sentinel errors and `FromErr` mapper. |
| `logger` | zerolog pre-configured with `Init(Config)` and level helpers. |
| `db/postgres` | GORM wrapper with tuneable pool, UTC `NowFunc` and prepared statement cache. |
| `cache/redis` | `go-redis/v9` wrapper covering strings, sets, TTL and ping. |
| `queue/nats` | `nats.go` wrapper with `PublishJSON` (sonic) and queue subscriptions. |
| `storage/s3` | `aws-sdk-go-v2/s3` wrapper for Put / Delete / Presign. Supports MinIO via endpoint override. |
| `httpclient/resty` | `go-resty/v2` with sonic, retries, fluent `Call` builder and JSON/form shortcuts. |
| `timeutil` | UTC-first helpers and `LoadLocation` with a configurable default. |
| `jsonutil` | sonic wrapper plus `Sanitize` / `SanitizeHeaders` for safe logging. |
| `crypto/aesgcm` | AES-256-GCM AEAD keyed from a base64 32-byte secret. |

## Identity middleware

The `middleware` package parses a set of `X-*` headers into a typed
`*Identity` made available through `FromContext`. The expected headers
are emitted by an upstream API gateway / auth service:

```
X-Is-Valid           true|false
X-Auth-Type          jwt|apikey|...
X-User-Uid           <uid>
X-Username           <string>
X-Email              <string>
X-Role               <string>
X-Is-Super-Admin     true|false
X-Is-Internal        true|false
X-User-Status        active|...
X-Timezone           <IANA tz>
X-Timezone-Offset    <±HH:MM>
X-Account-Uid        <uid>
X-Product-Slugs      slug1,slug2,...
X-Request-Id         <id>
```

Reject policy is configurable through `middleware.Options` — the default
returns 401 when `X-Is-Valid != "true"`, `X-User-Uid` is missing or
`X-Account-Uid` is missing.

## HTTP client (`httpclient/resty`)

`NewService` + `Call(ctx)` returns a fluent builder: chain `BearerToken` / `JSONDefaults` / `Header`… then a
terminal verb (`Get`, `Post`, …). `New` is the bare configured `resty.Client`. `GetRequest` and friends are
shortcuts for `Call` + defaults + verb; `*ExpectSuccess` helpers fail on non-2xx (`HTTPStatusCode`).

```go
import (
    "context"
    "os"

    xresty "github.com/danixts/platform/httpclient/resty"
)

ctx := context.Background()
api := xresty.NewService(xresty.Config{
    BaseURL: os.Getenv("UPSTREAM_API_URL"),
    Headers: map[string]string{"X-Api-Key": os.Getenv("API_KEY")},
})

var dto struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
if err := api.Call(ctx).BearerToken(os.Getenv("JWT")).JSONDefaults().Get("/v1/me", &dto); err != nil {
    return err
}
if err := api.GetRequestExpectSuccess(ctx, "/v1/me", &dto); err != nil {
    return err
}
```

## Quick start

```go
package main

import (
    "os"

    "github.com/gofiber/fiber/v3"

    xredis "github.com/danixts/platform/cache/redis"
    xpg    "github.com/danixts/platform/db/postgres"
    "github.com/danixts/platform/logger"
    xmw    "github.com/danixts/platform/middleware"
    xresp  "github.com/danixts/platform/response"
)

func main() {
    logger.Init(logger.Config{Level: "info", Service: "my-service"})

    db, err := xpg.New(xpg.Config{
        DSN:          os.Getenv("DATABASE_URL"),
        MaxOpenConns: 50,
        MaxIdleConns: 10,
    })
    if err != nil {
        logger.Fatal().Err(err).Msg("postgres")
    }

    rdb, err := xredis.New(xredis.Config{URL: os.Getenv("REDIS_URL")})
    if err != nil {
        logger.Fatal().Err(err).Msg("redis")
    }
    _ = db
    _ = rdb

    app := fiber.New()
    app.Use(xmw.RequestID(), xmw.Recover(), xmw.AccessLog())
    app.Use(xmw.Tenant())

    app.Get("/me", func(c fiber.Ctx) error {
        id := xmw.MustFromContext(c)
        return xresp.OK(c, id, "ok")
    })

    _ = app.Listen(":8080")
}
```

## Installation

```bash
go get github.com/danixts/platform@latest
```

## Versioning

Semantic versioning. `v0.x.y` is pre-stable — breaking changes may
happen in minor bumps until `v1.0.0`.

## License

MIT.
