package canonical

// TypingPayload is published to channel.typing.{session_uid} to signal
// that the bot is composing a response.
type TypingPayload struct {
	Typing     bool     `json:"typing"`
	SessionUID string   `json:"session_uid"`
	Provider   Provider `json:"provider"`
}
