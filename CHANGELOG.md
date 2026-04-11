# Changelog

All notable changes to platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial bootstrap of the SDK.
- `middleware` package: gateway identity middleware for Fiber v3 with
  `Tenant()` handler, `*Identity` struct, `FromContext` accessor and the
  `X-*` header contract emitted by an upstream auth gateway. Also ships
  `RequestID()`, `AccessLog()` and `Recover()` helpers.
- `response` package: standard `Body` envelope + generic `Page[T]` list
  envelope + sentinel errors + `FromErr` helper.
- `logger` package: zerolog wrapper with `Init(Config)` and level helpers.
- `db/postgres` package: GORM wrapper with tuneable pool, UTC NowFunc,
  prepared statement cache and translated errors.
- `cache/redis` package: go-redis/v9 wrapper covering strings, sets,
  TTL and readiness ping.
- `queue/nats` package: nats.go wrapper with `PublishJSON` (sonic) and
  queue subscribe helpers.
- `storage/s3` package: aws-sdk-go-v2/s3 wrapper for Put/Delete/PresignGet
  with MinIO-compatible endpoint override.
- `httpclient/resty` package: go-resty/v2 preconfigured with sonic,
  retry defaults and a browser-like User-Agent.
- `timeutil` package: UTC-first time helpers and `LoadLocation` with a
  configurable default location (`SetDefaultLocation`).
- `jsonutil` package: sonic wrapper + `Sanitize` / `SanitizeHeaders` for
  safe logging.
- `crypto/aesgcm` package: AES-256-GCM AEAD keyed from `APP_SECRET`.
