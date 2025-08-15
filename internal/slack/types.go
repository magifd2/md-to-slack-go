package slack

// SlackBlockKit is the top-level structure for the Slack Block Kit JSON.
type SlackBlockKit struct {
	Blocks []interface{} `json:"blocks"`
}

// HeaderBlock represents a header block.
type HeaderBlock struct {
	Type string    `json:"type"`
	Text TextBlock `json:"text"`
}

// SectionBlock represents a section block.
type SectionBlock struct {
	Type string    `json:"type"`
	Text TextBlock `json:"text"`
}

// DividerBlock represents a divider block.
type DividerBlock struct {
	Type string `json:"type"`
}

// TableBlock represents a table block.
type TableBlock struct {
	Type string             `json:"type"`
	Rows [][]RichTextObject `json:"rows"`
}

// TextBlock is a text object that can contain plain_text or mrkdwn.
type TextBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

// RichTextStyle represents the style of a text element.
type RichTextStyle struct {
	Bold bool `json:"bold,omitempty"`
}

// RichTextElement represents a text element within a rich text section.
type RichTextElement struct {
	Type  string         `json:"type"`
	Text  string         `json:"text"`
	Style *RichTextStyle `json:"style,omitempty"`
}

// RichTextSection represents a section within a rich text object.
type RichTextSection struct {
	Type     string            `json:"type"`
	Elements []RichTextElement `json:"elements"`
}

// RichTextObject represents a rich_text object used within table cells.
type RichTextObject struct {
	Type     string            `json:"type"`
	Elements []RichTextSection `json:"elements"`
}