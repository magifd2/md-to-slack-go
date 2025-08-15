package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/magifd2/md-to-slack-go/internal/slack"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// Convert converts a Markdown string to a Slack Block Kit JSON object.
func Convert(markdown string) (*slack.SlackBlockKit, error) {
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
				blocks = append(blocks, slack.HeaderBlock{Type: "header", Text: slack.TextBlock{Type: "plain_text", Text: string(heading.Text(source)), Emoji: true}})
			} else {
				blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: fmt.Sprintf("*%s*", string(heading.Text(source)))}})
			}
		case ast.KindParagraph:
			text := astToMrkdwn(node, source)
			if strings.TrimSpace(text) != "" {
				blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: text}})
			}
		case ast.KindBlockquote:
			quote := astToMrkdwn(node, source)
			lines := strings.Split(quote, "\n")
			for i := range lines {
				lines[i] = "> " + lines[i]
			}
			blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: strings.Join(lines, "\n")}})
		case ast.KindFencedCodeBlock:
			codeBlock := node.(*ast.FencedCodeBlock)
			lang := string(codeBlock.Info.Text(source))
			var codeLines []string
			for i := 0; i < codeBlock.Lines().Len(); i++ {
				line := codeBlock.Lines().At(i)
				codeLines = append(codeLines, string(line.Value(source)))
			}
			blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: fmt.Sprintf("```%s\n%s```", lang, strings.Join(codeLines, ""))}})
		case ast.KindThematicBreak:
			blocks = append(blocks, slack.DividerBlock{Type: "divider"})
		case extast.KindTable:
			tableNode := node.(*extast.Table)
			var tableRows [][]slack.RichTextObject

			// The first child of a Table is the Header row.
			headerRow := tableNode.FirstChild()
			if headerRow != nil {
				var headerCells []slack.RichTextObject
				for cell := headerRow.FirstChild(); cell != nil; cell = cell.NextSibling() {
					headerCells = append(headerCells, slack.CreateRichTextCell(string(cell.Text(source)), true))
				}
				tableRows = append(tableRows, headerCells)

				// Process the rest of the rows (the body).
				for rowNode := headerRow.NextSibling(); rowNode != nil; rowNode = rowNode.NextSibling() {
					var dataCells []slack.RichTextObject
					for cell := rowNode.FirstChild(); cell != nil; cell = cell.NextSibling() {
						dataCells = append(dataCells, slack.CreateRichTextCell(string(cell.Text(source)), false))
					}
					tableRows = append(tableRows, dataCells)
				}
			}

			blocks = append(blocks, slack.TableBlock{
				Type: "table",
				Rows: tableRows,
			})
		}
	}
	return &slack.SlackBlockKit{Blocks: blocks}, nil
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
