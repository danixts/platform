# canonical

Tipos y subjects NATS del bus de mensajería entre conectores y el core.

## Subjects

```go
canonical.SubjectChannelInbound    // "channel.inbound"
canonical.SubjectConnectorRawMeta  // "connector.raw.meta"
canonical.SubjectConnectorRawWA    // "connector.raw.wa"
canonical.SubjectOutboundMeta      // "channel.outbound.meta"
canonical.SubjectOutboundWA        // "channel.outbound.whatsmeow"
canonical.SubjectDeadletter        // "channel.deadletter"
canonical.SubjectTypingPrefix      // "channel.typing." + session_uid
```

## InboundEvent

Mensaje que llega de cualquier conector vía `channel.inbound`. El campo `Event` es discriminado por `type`:

```go
var e canonical.InboundEvent
sonic.Unmarshal(data, &e)

switch e.Event... // ver EventType constants
```

### Tipos de evento

```go
canonical.EventTypeMessage // "message" → MessageEvent { Text }
canonical.EventTypeAction  // "action"  → ActionEvent  { ActionID, Payload }
canonical.EventTypeMedia   // "media"   → MediaEvent   { MimeType, URL, Caption }
canonical.EventTypeStatus  // "status"  → StatusEvent  { Status, MessageID }
```

Unmarshal del evento interno:

```go
var msg canonical.MessageEvent
sonic.Unmarshal(e.Event, &msg)
```

## OutboundMessage

Mensaje saliente hacia un conector:

```go
out := canonical.OutboundMessage{
    SchemaVersion: canonical.CurrentSchemaVersion,
    MessageUID:    uid,
    SessionUID:    sessionUID,
    Provider:      canonical.ProviderMeta,
    Contact:       canonical.Contact{Phone: "+5491100000000"},
    Blocks:        blocks,
}
client.PublishJSON(canonical.SubjectOutboundMeta, out)
```

## Providers

```go
canonical.ProviderMeta      // "meta"
canonical.ProviderWhatsmeow // "whatsmeow"
```
