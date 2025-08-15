package markdown

import (
	"fmt"
	"os"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// DumpASTToStderr is a debugging function that walks the AST and prints its structure to stderr.
// This function is intended for development and debugging purposes and is not part of the main logic.
func DumpASTToStderr(doc ast.Node, source []byte) {
	var depth int

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			indent := strings.Repeat("    ", depth)
			fmt.Fprintf(os.Stderr, "%s%s (Type: %s)\n", indent, n.Kind(), n.Kind().String())

			// Print node-specific details
			switch node := n.(type) {
			case *ast.Text:
				fmt.Fprintf(os.Stderr, "%s  - Text: %q\n", indent, string(node.Text(source)))
			case *ast.Heading:
				fmt.Fprintf(os.Stderr, "%s  - Level: %d\n", indent, node.Level)
				fmt.Fprintf(os.Stderr, "%s  - Text: %q\n", indent, string(node.Text(source)))
			default:
				if textNode, ok := n.(interface{ Text([]byte) []byte }); ok {
					fmt.Fprintf(os.Stderr, "%s  - Text: %q\n", indent, string(textNode.Text(source)))
				}
			}

			// Lines() はブロックノードのみ
			if n.Type() == ast.TypeBlock && n.Lines() != nil && n.Lines().Len() > 0 {
				fmt.Fprintf(os.Stderr, "%s  - Lines:\n", indent)
				for i := 0; i < n.Lines().Len(); i++ {
					line := n.Lines().At(i)
					fmt.Fprintf(os.Stderr, "%s    - %q\n", indent, string(line.Value(source)))
				}
			}

			depth++
		} else {
			depth--
		}
		return ast.WalkContinue, nil
	})
}
