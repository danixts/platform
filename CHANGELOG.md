# Changelog

All notable changes to platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.1](https://github.com/danixts/platform/compare/v1.2.0...v1.2.1) (2026-04-24)


### Refactor

* **media:** introduce closeBody function for response body management ([044cad4](https://github.com/danixts/platform/commit/044cad4190645db84db64f183fa21ff559b0a7c2))

## [1.2.0](https://github.com/danixts/platform/compare/v1.1.0...v1.2.0) (2026-04-16)


### Features

* **middleware:** add OrgUID to Identity and X-Org-Uid header ([8e27aed](https://github.com/danixts/platform/commit/8e27aedc1b598d4ef88a36ece470b4ae9a7ec595))

## [1.1.0](https://github.com/danixts/platform/compare/v1.0.0...v1.1.0) (2026-04-16)


### Features

* **canonical:** add canonical channel schema package ([713f9a8](https://github.com/danixts/platform/commit/713f9a85bcd2ff28e98ea4f45759b280ae68bdca))


### Refactor

* **canonical:** update OutboundMessage struct to include Contact field ([b66a4a5](https://github.com/danixts/platform/commit/b66a4a5bafa824b6e277ee31018adcf5ec760f11))

## 1.0.0 (2026-04-12)


### Features

* **db/postgres:** add generic BaseRepo with Scope combinators ([11ed327](https://github.com/danixts/platform/commit/11ed327e96225f5fff2e66531f066af3462b960b))
* initial xmart-platform SDK bootstrap ([62311cf](https://github.com/danixts/platform/commit/62311cf28caa9ac51010699c9ab8adea4e0b7a45))
* rename module to platform, add requestid/accesslog/recover middlewares ([d461617](https://github.com/danixts/platform/commit/d4616179e39272cd0393e1b7eb75d3a5a64eb753))


### Refactor

* **db/postgres:** streamline BaseRepo and remove redundant comments ([6f84b94](https://github.com/danixts/platform/commit/6f84b94b3d4d8740cea17c7352c37d639af57cb0))

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
