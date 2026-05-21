package media

import "time"

type Config struct {
	BaseURL string
	Timeout time.Duration
	// SyncTimeout aplica solo a uploads con WaitForProcessed=true.
	// Si 0, se usa max(Timeout, 60s) para que el cliente nunca corte antes
	// que el server (MEDIA_SYNC_TIMEOUT default 30s) + margen de red.
	SyncTimeout time.Duration
}

type UploadOptions struct {
	AccountUID       string
	UserUID          string
	Bucket           string
	Service          string
	EntityType       string
	EntityID         string
	Preset           string
	Subpath          string
	Category         string
	Tags             []string
	WaitForProcessed bool
}

type VariantResp struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type MediaResp struct {
	UID      string `json:"uid"`
	URL      string `json:"url"`
	BlurHash string `json:"blurhash,omitempty"`
	// DominantColor: deprecated en favor de Palette[0]; se mantiene por compat.
	DominantColor string        `json:"dominant_color,omitempty"`
	Palette       []string      `json:"palette,omitempty"`
	Width         int           `json:"width,omitempty"`
	Height        int           `json:"height,omitempty"`
	Format        string        `json:"format,omitempty"`
	OriginalName  string        `json:"original_name"`
	ContentType   string        `json:"content_type"`
	SizeBytes     int64         `json:"size_bytes"`
	Preset        string        `json:"preset"`
	Status        string        `json:"status"`
	Tags          []string      `json:"tags,omitempty"`
	Category      string        `json:"category,omitempty"`
	Variants      []VariantResp `json:"variants,omitempty"`
}

type envelope struct {
	OK      bool      `json:"ok"`
	Message string    `json:"message"`
	Data    MediaResp `json:"data"`
}
