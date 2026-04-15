package canonical

const (
	SubjectChannelInbound   = "channel.inbound"
	SubjectConnectorRawMeta = "connector.raw.meta"
	SubjectConnectorRawWA   = "connector.raw.wa"
	SubjectOutboundMeta     = "channel.outbound.meta"
	SubjectOutboundWA       = "channel.outbound.whatsmeow"
	SubjectDeadletter       = "channel.deadletter"
	SubjectTypingPrefix     = "channel.typing." // append session_uid
)
