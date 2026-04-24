# timeutil

Helpers de tiempo. Todas las funciones retornan UTC.

## Funciones

```go
timeutil.Now()            // time.Time en UTC
timeutil.NowRFC3339()     // "2026-04-24T17:00:00Z"
timeutil.NowUnix()        // unix timestamp int64
timeutil.Add(2*time.Hour) // Now() + duración

timeutil.Unix(ts)         // int64 → time.Time UTC
timeutil.UTC(t)           // convierte cualquier time.Time a UTC
timeutil.Duration(60)     // 60 → time.Duration (segundos)

timeutil.ParseRFC3339(s)  // string → time.Time UTC, error si inválido
```

## Zona horaria por defecto

Configurable una vez en startup (default: UTC):

```go
loc, _ := time.LoadLocation("America/Buenos_Aires")
timeutil.SetDefaultLocation(loc)

// en helpers que respetan la zona:
timeutil.LoadLocation("America/Buenos_Aires") // loc o DefaultLocation() si falla
timeutil.LoadLocation("")                     // retorna DefaultLocation()
```
