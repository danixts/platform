# storage/s3

Cliente S3 compatible (AWS S3, MinIO, R2).

## Conexión

```go
client, err := s3.New(ctx, s3.Config{
    Region:         "us-east-1",
    Bucket:         "my-bucket",
    AccessKey:      "...",
    SecretKey:      "...",
    Endpoint:       "https://minio.example.com", // dejar vacío para AWS nativo
    ForcePathStyle: true,                         // requerido para MinIO
    Prefix:         "uploads",                    // prefijo global de keys
    CDNBaseURL:     "https://cdn.example.com",    // si vacío usa URL de S3
    ACLPublic:      true,
})
```

## Subir

```go
url, err := client.Put(ctx, "avatars/user-123.jpg", imageBytes, "image/jpeg")
// url = "https://cdn.example.com/uploads/avatars/user-123.jpg"
```

## Eliminar

```go
err := client.Delete(ctx, "avatars/user-123.jpg")
```

## URL firmada (acceso temporal)

```go
url, err := client.PresignGet(ctx, "docs/contract.pdf", 1*time.Hour)
```

## Helpers de key

```go
fullKey := client.JoinKey("avatars/user-123.jpg") // "uploads/avatars/user-123.jpg"
pubURL  := client.PublicURL(fullKey)
bucket  := client.Bucket()
```
