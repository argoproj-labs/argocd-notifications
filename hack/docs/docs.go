package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/argoproj-labs/argocd-notifications/builtin"

	"github.com/olekukonko/tablewriter"
)

func builtInDocs(out io.Writer) {
	_, _ = fmt.Fprintln(out, "# Built-in Triggers and Templates")
	_, _ = fmt.Fprintln(out, "## Triggers")

	triggers := tablewriter.NewWriter(out)
	triggers.SetHeader([]string{"NAME", "DESCRIPTION", "TEMPLATE"})
	triggers.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	triggers.SetCenterSeparator("|")
	triggers.SetAutoWrapText(false)
	for _, t := range builtin.Triggers {
		triggers.Append([]string{t.Name, t.Description, fmt.Sprintf("[%s](#%s)", t.Template, t.Template)})
	}
	triggers.Render()

	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintln(out, "## Templates")
	for _, t := range builtin.Templates {
		_, _ = fmt.Fprintf(out, "### %s\n**title**: `%s`\n\n**body**:\n```\n%s\n```\n", t.Name, t.Title, t.Body)
	}
}

func main() {
	var builtItDocsData bytes.Buffer
	builtInDocs(&builtItDocsData)
	if err := ioutil.WriteFile("./docs/built-in.md", builtItDocsData.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
