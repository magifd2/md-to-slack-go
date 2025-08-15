package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

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

// RichTextObject represents a rich_text object used within table cells.
type RichTextObject struct {
	Type     string `json:"type"`
	Elements []struct {
		Type     string `json:"type"`
		Elements []struct {
			Type  string `json:"type"`
			Text  string `json:"text"`
			Style *struct {
				Bold bool `json:"bold,omitempty"`
			} `json:"style,omitempty"`
		} `json:"elements"`
	} `json:"elements"`
}

// createRichTextCell generates a rich_text cell object for Slack table blocks.
func createRichTextCell(content string, isHeader bool) RichTextObject {
	// If the cell text is empty, insert a space to avoid API errors.
	if strings.TrimSpace(content) == "" {
		content = " "
	}

	cell := RichTextObject{
		Type: "rich_text",
		Elements: []struct {
			Type     string `json:"type"`
			Elements []struct {
				Type  string `json:"type"`
				Text  string `json:"text"`
				Style *struct {
					Bold bool `json:"bold,omitempty"`
				} `json:"style,omitempty"`
			} `json:"elements"`
		}{
			{
				Type: "rich_text_section",
				Elements: []struct {
					Type  string `json:"type"`
					Text  string `json:"text"`
					Style *struct {
						Bold bool `json:"bold,omitempty"`
					} `json:"style,omitempty"`
				}{
					{
						Type: "text",
						Text: content,
					},
				},
			},
		},
	}
	if isHeader {
		cell.Elements[0].Elements[0].Style = &struct {
			Bold bool `json:"bold,omitempty"`
		}{Bold: true}
	}
	return cell
}

// astToMrkdwn recursively traverses AST nodes and converts them to a Slack mrkdwn formatted string.
func astToMrkdwn(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case ast.KindText:
			buf.WriteString(string(c.(*ast.Text).Text(source)))
		case ast.KindString:
			buf.WriteString(string(c.(*ast.String).Value))
		case ast.KindCodeSpan:
			buf.WriteString("`")
			buf.WriteString(string(c.(*ast.CodeSpan).Text(source)))
			buf.WriteString("`")
		case ast.KindEmphasis:
			em := c.(*ast.Emphasis)
			marker := "_"
			if em.Level == 2 {
				marker = "*"
			}
			buf.WriteString(marker)
			buf.WriteString(astToMrkdwn(c, source))
			buf.WriteString(marker)
		case ast.KindLink:
			link := c.(*ast.Link)
			buf.WriteString(fmt.Sprintf("<%s|%s>", link.Destination, astToMrkdwn(c, source)))
		case ast.KindImage:
			img := c.(*ast.Image)
			buf.WriteString(fmt.Sprintf("<%s|%s>", img.Destination, astToMrkdwn(c, source)))
		case ast.KindAutoLink:
			link := c.(*ast.AutoLink)
			url := string(link.URL(source))
			buf.WriteString(fmt.Sprintf("<%s|%s>", url, url))
		case extast.KindStrikethrough:
			buf.WriteString("~")
			buf.WriteString(astToMrkdwn(c, source))
			buf.WriteString("~")
		default:
			buf.WriteString(astToMrkdwn(c, source))
		}
	}
	return buf.String()
}

// markdownToSlackBlocks converts a Markdown string to a Slack Block Kit JSON object.
func markdownToSlackBlocks(markdown string) (*SlackBlockKit, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)
	source := []byte(markdown)
	doc := md.Parser().Parse(text.NewReader(source))

	var blocks []interface{}
	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		switch node.Kind() {
		case ast.KindHeading:
			heading := node.(*ast.Heading)
			if heading.Level <= 2 {
				blocks = append(blocks, HeaderBlock{Type: "header", Text: TextBlock{Type: "plain_text", Text: string(heading.Text(source)), Emoji: true}})
			} else {
				blocks = append(blocks, SectionBlock{Type: "section", Text: TextBlock{Type: "mrkdwn", Text: fmt.Sprintf("*%s*", string(heading.Text(source)))}})
			}
		case ast.KindParagraph:
			text := astToMrkdwn(node, source)
			if strings.TrimSpace(text) != "" {
				blocks = append(blocks, SectionBlock{Type: "section", Text: TextBlock{Type: "mrkdwn", Text: text}})
			}
		case ast.KindBlockquote:
			quote := astToMrkdwn(node, source)
			lines := strings.Split(quote, "\n")
			for i := range lines {
				lines[i] = "> " + lines[i]
			}
			blocks = append(blocks, SectionBlock{Type: "section", Text: TextBlock{Type: "mrkdwn", Text: strings.Join(lines, "\n")}})
		case ast.KindFencedCodeBlock:
			codeBlock := node.(*ast.FencedCodeBlock)
			lang := string(codeBlock.Info.Text(source))
			var codeLines []string
			for i := 0; i < codeBlock.Lines().Len(); i++ {
				line := codeBlock.Lines().At(i)
				codeLines = append(codeLines, string(line.Value(source)))
			}
			blocks = append(blocks, SectionBlock{Type: "section", Text: TextBlock{Type: "mrkdwn", Text: fmt.Sprintf("```%s\n%s```", lang, strings.Join(codeLines, ""))}})
		case ast.KindThematicBreak:
			blocks = append(blocks, DividerBlock{Type: "divider"})
		case extast.KindTable:
			tableNode := node.(*extast.Table)
			var tableRows [][]RichTextObject

			// The first child of a Table is the Header row.
			headerRow := tableNode.FirstChild()
			if headerRow != nil {
				var headerCells []RichTextObject
				for cell := headerRow.FirstChild(); cell != nil; cell = cell.NextSibling() {
					headerCells = append(headerCells, createRichTextCell(string(cell.Text(source)), true))
				}
				tableRows = append(tableRows, headerCells)

				// Process the rest of the rows (the body).
				for rowNode := headerRow.NextSibling(); rowNode != nil; rowNode = rowNode.NextSibling() {
					var dataCells []RichTextObject
					for cell := rowNode.FirstChild(); cell != nil; cell = cell.NextSibling() {
						dataCells = append(dataCells, createRichTextCell(string(cell.Text(source)), false))
					}
					tableRows = append(tableRows, dataCells)
				}
			}

			blocks = append(blocks, TableBlock{
				Type: "table",
				Rows: tableRows,
			})
		}
	}
	return &SlackBlockKit{Blocks: blocks}, nil
}

func main() {
	flag.Parse()
	var input []byte
	var err error
	if args := flag.Args(); len(args) > 0 {
		input, err = os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	}
	if strings.TrimSpace(string(input)) == "" {
		return
	}

	slackBlocks, err := markdownToSlackBlocks(string(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(slackBlocks, "", "  ")
	fmt.Println(string(out))
}
