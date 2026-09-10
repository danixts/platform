package canonical

import "encoding/json"

// InboundEvent is the canonical message arriving from any connector via channel.inbound.
// Callers unmarshal Event into one of the typed event structs based on the "type" field.
//
// Example:
//
//	var e canonical.InboundEvent
//	json.Unmarshal(data, &e)
//	var msg canonical.MessageEvent
//	json.Unmarshal(e.Event, &msg)
type InboundEvent struct {
	SchemaVersion string          `json:"schema_version"`
	MessageUID    string          `json:"message_uid"` // UUID v7
	Provider      Provider        `json:"provider"`
	Channel       string          `json:"channel,omitempty"` // sub-channel: whatsapp|messenger|instagram
	SessionUID    string          `json:"session_uid"`
	Contact       Contact         `json:"contact"`
	ReceivedAt    string          `json:"received_at"` // ISO 8601 UTC
	Event         json.RawMessage `json:"event"`       // discriminated by event.type
}

// EventType constants.
const (
	EventTypeMessage      = "message"
	EventTypeAction       = "action"
	EventTypeMedia        = "media"
	EventTypeStatus       = "status"
	EventTypeFlowResponse = "flow_response"
)

// MessageEvent represents a plain-text or rich-text inbound message.
type MessageEvent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ActionEvent represents an interactive action (e.g. button tap).
type ActionEvent struct {
	Type     string `json:"type"`
	ActionID string `json:"action_id"` // format: {block_uid}:{option_id}
	Payload  string `json:"payload,omitempty"`
}

// MediaEvent represents an inbound media attachment.
type MediaEvent struct {
	Type     string `json:"type"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

// StatusEvent represents a delivery or read receipt.
type StatusEvent struct {
	Type      string `json:"type"`
	Status    string `json:"status"` // "delivered" | "read" | "failed"
	MessageID string `json:"message_id"`
}

// FlowResponseEvent represents a completed WhatsApp Flow submission.
// CorrelationToken echoes the one the connector sent in the outbound
// FlowBlock, so the bot can resolve which conversation and node was waiting
// for it. Fields holds the flow's screen data already parsed by the
// connector; RawResponse keeps the untouched payload for debugging.
type FlowResponseEvent struct {
	Type             string          `json:"type"`
	CorrelationToken string          `json:"correlation_token"`
	FlowID           string          `json:"flow_id"`
	Fields           map[string]any  `json:"fields"`
	RawResponse      json.RawMessage `json:"raw_response,omitempty"`
}
