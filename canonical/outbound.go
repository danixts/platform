package canonical

type OutboundMessage struct {
	SchemaVersion    string   `json:"schema_version"`
	MessageUID       string   `json:"message_uid"`
	SessionUID       string   `json:"session_uid"`
	Provider         Provider `json:"provider"`
	Contact          Contact  `json:"contact"`
	Blocks           []Block  `json:"blocks"`
	RequiresTemplate bool     `json:"requires_template"`
	LastInboundAt    string   `json:"last_inbound_at,omitempty"`
	TemplateName     string   `json:"template_name,omitempty"`
	TemplateLanguage string   `json:"template_language,omitempty"`
	FallbackText     string   `json:"fallback_text,omitempty"`
}
