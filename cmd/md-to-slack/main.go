package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/magifd2/md-to-slack-go/internal/markdown"
)

var version = "dev" // This will be overwritten by LDFLAGS

func main() {
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

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

	slackBlocks, err := markdown.Convert(string(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(slackBlocks, "", "  ")
	fmt.Println(string(out))
}
