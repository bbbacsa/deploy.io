package main

import (
	"flag"
	"fmt"
	"github.com/bbbacsa/deploy.io/commands"
	"github.com/bbbacsa/deploy.io/constants"
	"io"
	"os"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("Deploy %s\n", constants.Version)
		return
	}

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	for _, cmd := range commands.All {
		if cmd.Name() == args[0] {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			err := cmd.Run(cmd, args)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown subcommand: %q\n\n", args[0])
	usage()
}

var usageTemplate = `Deploy.IO command-line client.

Usage: deploy COMMAND [ARG...]

Commands:
{{range .}}
  {{.Name | printf "%-11s"}} {{.Short}}{{end}}

Run 'deploy COMMAND -h' for more information on a command.
`

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, commands.All)
}

func usage() {
	printUsage(os.Stderr)
	os.Exit(2)
}

