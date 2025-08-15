package slack

import "strings"

// CreateRichTextCell generates a rich_text cell object for Slack table blocks.
func CreateRichTextCell(content string, isHeader bool) RichTextObject {
	if strings.TrimSpace(content) == "" {
		content = " "
	}

	textElement := RichTextElement{
		Type: "text",
		Text: content,
	}

	if isHeader {
		textElement.Style = &RichTextStyle{Bold: true}
	}

	sectionElement := RichTextSection{
		Type:     "rich_text_section",
		Elements: []RichTextElement{textElement},
	}

	cell := RichTextObject{
		Type:     "rich_text",
		Elements: []RichTextSection{sectionElement},
	}

	return cell
}
