# media

Cliente HTTP para el servicio de media interno.

## Setup

```go
client := media.New(media.Config{
    BaseURL: "https://media.internal",
    Timeout: 15 * time.Second,
})
```

## Subir archivo

```go
f, _ := os.Open("photo.jpg")
defer f.Close()

result, err := client.Upload(ctx, "photo.jpg", f, "image/jpeg", media.UploadOptions{
    AccountUID: id.AccountUID,
    Bucket:     "avatars",
    Service:    "user-service",
    EntityType: "user",
    EntityID:   userUID,
    Preset:     "avatar",
    Tags:       []string{"profile"},
})

fmt.Println(result.URL)
```

## Obtener metadata

```go
m, err := client.Get(ctx, accountUID, mediaUID)
if m == nil { /* no encontrado */ }
```

## Eliminar

```go
err := client.Delete(ctx, accountUID, mediaUID)
```

## MediaResp

```go
type MediaResp struct {
    UID           string
    URL           string
    BlurHash      string
    DominantColor string
    Width, Height int
    Format        string
    OriginalName  string
    ContentType   string
    SizeBytes     int64
    Preset        string
    Status        string
    Tags          []string
    Category      string
    Variants      []VariantResp // { Name, URL, Width, Height }
}
```
