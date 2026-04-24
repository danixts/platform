# queue/nats

Cliente NATS con soporte completo de pub/sub y req/reply tipado.

## Conexión

```go
client, err := nats.New(nats.Config{
    URL:        "nats://localhost:4222",
    Name:       "my-service",
    QueueGroup: "my-service-group", // opcional; activa queue subscribe automático
})
```

## Publicar

```go
// raw
client.Publish("user.created", data)

// JSON
client.PublishJSON("user.created", UserCreatedEvent{UserUID: "..."})
```

## Suscribir (fire-and-forget)

```go
client.Subscribe("user.created", func(subject string, data []byte) {
    // procesar
})
```

### Suscribir con unmarshal automático

```go
nats.SubscribeJSON[UserCreatedEvent](client, "user.created", func(msg *gonats.Msg, e UserCreatedEvent) {
    fmt.Println(e.UserUID)
    // msg disponible si necesitás responder
})
```

## Req/Reply — servidor

```go
nats.SubscribeJSON[CreateUserReq](client, "user.create", func(msg *gonats.Msg, req CreateUserReq) {
    user, err := uc.Create(ctx, req)
    if err != nil {
        nats.Reply(msg, ErrorResp{Error: err.Error()})
        return
    }
    nats.Reply(msg, user)
})
```

## Req/Reply — cliente

Tipado con genéricos:

```go
user, err := nats.Call[CreateUserReq, UserResp](ctx, client, "user.create", CreateUserReq{
    Name:  "John",
    Email: "john@example.com",
})
```

Raw (sin tipado):

```go
resp, err := client.RequestJSON(ctx, "user.create", req, &out)
```

## Timeout

El timeout lo controla el contexto del llamador:

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

user, err := nats.Call[CreateUserReq, UserResp](ctx, client, "user.create", req)
```

## Tipos de handler

| Tipo | Uso |
|------|-----|
| `Handler` | fire-and-forget, solo subject + bytes |
| `MsgHandler` | acceso al `*nats.Msg` completo (req/reply) |
| `SubscribeJSON[T]` | unmarshal automático + acceso al msg |
