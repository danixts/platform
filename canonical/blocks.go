package canonical

import (
	"encoding/json"
	"fmt"
)

// Block is a json.RawMessage — callers unmarshal based on the "type" field.
// Use BlockType* constants to discriminate.
type Block = json.RawMessage

const (
	BlockTypeText            = "text"
	BlockTypeImage           = "image"
	BlockTypeVideo           = "video"
	BlockTypeAudio           = "audio"
	BlockTypeFile            = "file"
	BlockTypeButtons         = "buttons"
	BlockTypeListPicker      = "list_picker"
	BlockTypeProductCarousel = "product_carousel"
	BlockTypeCTAURL          = "cta_url"
	BlockTypeQuickReplies    = "quick_replies"
	BlockTypeForm            = "form"
)

// TextBlock renders a text message.
type TextBlock struct {
	Type       string `json:"type"`               // "text"
	Content    string `json:"content"`
	Formatting string `json:"formatting,omitempty"` // "markdown" | "plain"
}

// ImageBlock renders an image attachment.
type ImageBlock struct {
	Type     string `json:"type"` // "image"
	URL      string `json:"url"`
	Caption  string `json:"caption,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// ButtonOption is a single interactive button.
type ButtonOption struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Action  string `json:"action,omitempty"`
	Payload string `json:"payload,omitempty"`
}

// ButtonsBlock renders a set of interactive buttons with an optional prompt.
type ButtonsBlock struct {
	Type    string         `json:"type"` // "buttons"
	Prompt  string         `json:"prompt"`
	Options []ButtonOption `json:"options"`
}

// ListItem is a single row in a list picker.
type ListItem struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// ListPickerBlock renders a scrollable list of selectable items.
type ListPickerBlock struct {
	Type  string     `json:"type"` // "list_picker"
	Title string     `json:"title"`
	Items []ListItem `json:"items"`
}

// QuickReplyOption is a single quick-reply chip.
type QuickReplyOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// QuickRepliesBlock renders a row of quick-reply chips.
type QuickRepliesBlock struct {
	Type    string             `json:"type"` // "quick_replies"
	Options []QuickReplyOption `json:"options"`
}

// CTAURLBlock renders a call-to-action button that opens a URL.
type CTAURLBlock struct {
	Type  string `json:"type"` // "cta_url"
	Label string `json:"label"`
	URL   string `json:"url"`
}

// MustMarshalBlock marshals any typed block to json.RawMessage for use in OutboundMessage.Blocks.
// Panics if marshaling fails (indicates a programming error with non-serializable types).
func MustMarshalBlock(v any) Block {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("canonical: MustMarshalBlock: %v", err))
	}
	return b
}

// BlockTypeOf returns the "type" field from a raw Block without a full unmarshal.
func BlockTypeOf(b Block) string {
	var m struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(b, &m)
	return m.Type
}
