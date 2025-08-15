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
		switch n := node.(type) {
		case *ast.Heading:
			if n.Level <= 2 {
				blocks = append(blocks, slack.HeaderBlock{Type: "header", Text: slack.TextBlock{Type: "plain_text", Text: string(n.Text(source)), Emoji: true}})
			} else {
				blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: fmt.Sprintf("*%s*", string(n.Text(source)))}})
			}
		case *ast.Paragraph:
			if n.ChildCount() == 1 && n.FirstChild().Kind() == ast.KindImage {
				img := n.FirstChild().(*ast.Image)
				blocks = append(blocks, slack.ImageBlock{
					Type:     "image",
					ImageURL: string(img.Destination),
					AltText:  string(img.Text(source)),
				})
			} else {
				text := astToMrkdwn(n, source)
				if strings.TrimSpace(text) != "" {
					blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: text}})
				}
			}
		case *ast.Blockquote:
			quote := astToMrkdwn(n, source)
			lines := strings.Split(quote, "\n")
			var quoteText strings.Builder
			for i, line := range lines {
				quoteText.WriteString("> ")
				quoteText.WriteString(line)
				if i < len(lines)-1 {
					quoteText.WriteString("\n")
				}
			}
			blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: quoteText.String()}})
		case *ast.FencedCodeBlock:
			lang := string(n.Info.Text(source))
			var codeLines []string
			for i := 0; i < n.Lines().Len(); i++ {
				line := n.Lines().At(i)
				codeLines = append(codeLines, string(line.Value(source)))
			}
			blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: fmt.Sprintf("```%s\n%s```", lang, strings.Join(codeLines, ""))}})
		case *ast.ThematicBreak:
			blocks = append(blocks, slack.DividerBlock{Type: "divider"})
		case *extast.Table:
			var tableRows [][]slack.RichTextObject
			header := n.FirstChild()
			if header != nil {
				var headerCells []slack.RichTextObject
				for cell := header.FirstChild(); cell != nil; cell = cell.NextSibling() {
					headerCells = append(headerCells, slack.CreateRichTextCell(string(cell.Text(source)), true))
				}
			tableRows = append(tableRows, headerCells)
				for row := header.NextSibling(); row != nil; row = row.NextSibling() {
					var rowCells []slack.RichTextObject
					for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
						rowCells = append(rowCells, slack.CreateRichTextCell(string(cell.Text(source)), false))
					}
					tableRows = append(tableRows, rowCells)
				}
			}
			blocks = append(blocks, slack.TableBlock{Type: "table", Rows: tableRows})
		case *ast.List:
			listText := listToMrkdwn(n, source, 0)
			blocks = append(blocks, slack.SectionBlock{Type: "section", Text: slack.TextBlock{Type: "mrkdwn", Text: listText}})
		}
	}
	return &slack.SlackBlockKit{Blocks: blocks}, nil
}

func listToMrkdwn(l *ast.List, source []byte, depth int) string {
	var buf bytes.Buffer
	indent := strings.Repeat("  ", depth)

	for c := l.FirstChild(); c != nil; c = c.NextSibling() {
		item := c.(*ast.ListItem)

		marker := "- "
		if l.IsOrdered() {
			i := 0
			for p := c.PreviousSibling(); p != nil; p = p.PreviousSibling() {
				i++
			}
			marker = fmt.Sprintf("%d. ", l.Start+i)
		}

		buf.WriteString(indent)
		buf.WriteString(marker)

		var contentBuf bytes.Buffer
		for child := item.FirstChild(); child != nil; child = child.NextSibling() {
			if child.Kind() == ast.KindList {
				contentBuf.WriteString("\n")
				contentBuf.WriteString(listToMrkdwn(child.(*ast.List), source, depth+1))
			} else {
				contentBuf.WriteString(astToMrkdwn(child, source))
			}
		}
		buf.WriteString(contentBuf.String())
		buf.WriteString("\n")
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

func astToMrkdwn(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		kind := c.Kind().String()
		if kind == "HardLineBreak" || kind == "SoftLineBreak" {
			buf.WriteString("\n")
			continue
		}

		switch n := c.(type) {
		case *ast.Text:
			buf.WriteString(string(n.Text(source)))
		case *ast.String:
			buf.WriteString(string(n.Value))
		case *ast.CodeSpan:
			buf.WriteString("`")
			buf.WriteString(string(n.Text(source)))
			buf.WriteString("`")
		case *ast.Emphasis:
			marker := "_"
			if n.Level == 2 {
				marker = "*"
			}
			buf.WriteString(marker)
			buf.WriteString(astToMrkdwn(n, source))
			buf.WriteString(marker)
		case *ast.Link:
			buf.WriteString(fmt.Sprintf("<%s|%s>", n.Destination, astToMrkdwn(n, source)))
		case *ast.Image:
			buf.WriteString(fmt.Sprintf("<%s|%s>", n.Destination, astToMrkdwn(n, source)))
		case *ast.AutoLink:
			url := string(n.URL(source))
			buf.WriteString(fmt.Sprintf("<%s|%s>", url, url))
		case *extast.Strikethrough:
			buf.WriteString("~")
			buf.WriteString(astToMrkdwn(n, source))
			buf.WriteString("~")
		default:
			buf.WriteString(astToMrkdwn(c, source))
		}
	}
	return buf.String()
}