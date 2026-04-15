package canonical

// Provider identifies the messaging channel connector.
type Provider string

const (
	ProviderMeta      Provider = "meta"
	ProviderWhatsmeow Provider = "whatsmeow"
)

// Contact holds basic sender/recipient information.
type Contact struct {
	Phone string `json:"phone"`
	Name  string `json:"name,omitempty"`
}

// SchemaVersion is the version string type for canonical messages.
type SchemaVersion = string

// CurrentSchemaVersion is the schema version implemented by this package.
const CurrentSchemaVersion SchemaVersion = "0.1.0"
