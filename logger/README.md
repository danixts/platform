# logger

Logger global basado en zerolog. Inicializar una vez en `main`, usar desde cualquier paquete.

## Init (en main)

```go
logger.Init(logger.Config{
    Level:   "info",   // trace | debug | info | warn | error | fatal
    Pretty:  false,    // true para desarrollo local
    Service: "my-service",
})
```

## Uso

```go
logger.Info().Str("user_uid", uid).Msg("user created")
logger.Warn().Str("subject", s).Msg("nats handler slow")
logger.Error().Err(err).Str("path", c.Path()).Msg("handler failed")
logger.Fatal().Err(err).Msg("startup failed")
```

## Logger base

Para pasarlo a librerías que aceptan `*zerolog.Logger`:

```go
log := logger.Get()
```

## Niveles

| Función | Uso |
|---------|-----|
| `Trace()` | diagnóstico muy detallado |
| `Debug()` | desarrollo local |
| `Info()` | eventos normales del sistema |
| `Warn()` | condiciones anómalas no fatales |
| `Error()` | errores manejados |
| `Fatal()` | error irrecuperable, termina el proceso |
