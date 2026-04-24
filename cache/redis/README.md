# cache/redis

Cliente Redis con API simplificada sobre go-redis.

## Conexión

```go
client, err := redis.New(redis.Config{
    URL:         "redis://localhost:6379",
    PingTimeout: 3 * time.Second, // default: 3s
})
```

## Operaciones básicas

```go
// set con TTL
client.Set(ctx, "session:abc", token, 24*time.Hour)

// get
value, found, err := client.Get(ctx, "session:abc")
if !found { /* cache miss */ }

// delete
client.Delete(ctx, "session:abc", "session:xyz")

// existe
ok, err := client.Exists(ctx, "session:abc")

// extender TTL
client.Expire(ctx, "session:abc", 1*time.Hour)
```

## Sets

```go
client.SAdd(ctx, "org:123:members", "user:a", "user:b")
client.SRem(ctx, "org:123:members", "user:a")

members, err := client.SMembers(ctx, "org:123:members")
ok, err := client.SIsMember(ctx, "org:123:members", "user:b")
```

## Acceso al cliente subyacente

Para operaciones no expuestas en la API:

```go
rdb := client.Raw() // *goredis.Client
```
