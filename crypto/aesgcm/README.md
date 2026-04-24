# crypto/aesgcm

Cifrado simétrico AES-256-GCM. Output en base64 URL-safe sin padding.

## Setup

La clave debe ser exactamente 32 bytes codificados en base64 estándar:

```bash
# generar clave
openssl rand -base64 32
```

```go
cipher, err := aesgcm.New(os.Getenv("APP_SECRET"))
```

## Cifrar / Descifrar

```go
encrypted, err := cipher.EncryptString("valor sensible")
plain, err     := cipher.DecryptString(encrypted)

// con bytes
enc, err  := cipher.Encrypt([]byte{...})
dec, err  := cipher.Decrypt(enc)
```

## Placeholder

Para detectar configs sin inicializar en producción:

```go
if aesgcm.IsPlaceholder(val) {
    // config no configurada
}
```

El valor placeholder es `<<REPLACE_WITH_ENCRYPTED>>`. `Decrypt` retorna `ErrPlaceholderSecret` si lo recibe.

## Errores

| Error | Causa |
|-------|-------|
| `ErrInvalidKey` | clave no es base64 de 32 bytes |
| `ErrCipherTooShort` | ciphertext corrupto o truncado |
| `ErrPlaceholderSecret` | se intentó descifrar el placeholder |
