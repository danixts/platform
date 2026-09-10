# jsonutil

Utilidades para sanitización y parsing seguro de JSON.

## Sanitizar payload antes de loggear

Redacta automáticamente campos sensibles en estructuras `map[string]any`:

```go
raw := jsonutil.TryUnmarshal(body)
safe := jsonutil.Sanitize(raw)
logger.Info().Interface("body", safe).Msg("request received")
```

Campos redactados por defecto: `password`, `apiKey`, `api_key`, `Authorization`, `token`, `access_token`, `refresh_token`, `secret`, `client_secret`, `private_key`, entre otros.

## Sanitizar headers

```go
safe := jsonutil.SanitizeHeaders(map[string]string{
    "Authorization": "Bearer abc123",
    "Content-Type":  "application/json",
})
// Authorization → "[REDACTED]"
```

## TryUnmarshal

Parsea bytes sin fallar: si el JSON es inválido devuelve `{"raw": "<string>"}`:

```go
v := jsonutil.TryUnmarshal(data)
```

## Agregar clave sensible

```go
jsonutil.SensitiveKeys["my_secret_field"] = true
```
