# platform

SDK interno de XMart Cloud. Wrappers finos sobre clientes de infraestructura
(Postgres, Redis, NATS, S3, Resty) y utilitarios compartidos (middleware de
identidad, logger, response envelope, crypto, timeutil).

## Principios

- **Wrappers finos, no frameworks**. Cada cliente expone `New(Config)` y un
  puñado de métodos de alto nivel. Cuando el wrapper no alcanza, `Raw()`
  devuelve el cliente subyacente.
- **Context-first**. Todo I/O recibe `context.Context` como primer parámetro.
- **Sin panics en producción**. Las únicas excepciones son errores de
  programación (p.ej. usar `middleware.FromContext` sin el middleware en el chain).
- **Cero dependencia de frameworks salvo Fiber v3** para el middleware HTTP.
- **Compatibilidad semver**. Breaking changes sólo en major bumps.

## Paquetes

| Paquete | Descripción |
|---|---|
| `middleware` | Identity unificado: parsea headers `X-*` del gateway y expone `*Identity` request-scoped. |
| `response` | Envelope estándar `Body[T]` y paginado `Page[T]`. |
| `logger` | Zerolog pre-configurado con salida JSON en prod y pretty en dev. |
| `db/postgres` | Wrapper GORM con pool tuneable y logger integrado. |
| `cache/redis` | Wrapper `go-redis/v9` con helpers de strings, sets y rate-limit. |
| `queue/nats` | Wrapper `nats.go` con `PublishJSON` y `QueueSubscribe`. |
| `storage/s3` | Wrapper `aws-sdk-go-v2/s3` para Put/Get/Delete/Presign (soporta MinIO via endpoint). |
| `httpclient/resty` | Resty v2 pre-configurado con sonic, retry y UA default. |
| `timeutil` | `LoadLocation` con fallback a `America/La_Paz`. |
| `jsonutil` | Wrapper sonic + `SanitizeHeaders` para logs. |
| `crypto/aesgcm` | AES-256-GCM con clave derivada de `APP_SECRET`. |
| `pagination` | Parseo de query params `page`/`page_size`. |

## Uso

```go
import (
    xmw   "github.com/danixts/platform/middleware"
    xpg   "github.com/danixts/platform/db/postgres"
    xresp "github.com/danixts/platform/response"
)

func main() {
    db, err := xpg.New(xpg.Config{
        DSN:          os.Getenv("DATABASE_URL"),
        MaxOpenConns: 50,
        MaxIdleConns: 10,
    })
    // ...

    app := fiber.New()
    app.Use(xmw.Tenant())

    app.Get("/me", func(c fiber.Ctx) error {
        id, err := xmw.FromContext(c)
        if err != nil {
            return xresp.Unauthorized(c, "missing_identity")
        }
        return xresp.OK(c, id)
    })
}
```

## Consumiendo este módulo privado

```bash
export GOPRIVATE=github.com/danixts/*
git config --global url."git@github.com:danixts/".insteadOf "https://github.com/danixts/"
go get github.com/danixts/platform@v0.1.0
```

En CI (GitHub Actions):

```yaml
- name: Configure Go private modules
  env:
    GOPRIVATE: github.com/danixts/*
    GH_TOKEN: ${{ secrets.GH_PRIVATE_MODULES_TOKEN }}
  run: |
    git config --global url."https://x-access-token:${GH_TOKEN}@github.com/danixts/".insteadOf "https://github.com/danixts/"
```

## Versionado

Semver estricto. `v0.x.y` es pre-estable — breaking changes pueden suceder
en minor bumps hasta `v1.0.0`.
