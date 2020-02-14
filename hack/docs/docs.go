package main

import (
	"fmt"
	"io"
	"os"

	"github.com/argoproj-labs/argocd-notifications/builtin"

	"github.com/olekukonko/tablewriter"
)

func generate(out io.Writer) {
	fmt.Println("# Built-in Triggers and Templates")
	fmt.Println("## Triggers")

	triggers := tablewriter.NewWriter(out)
	triggers.SetHeader([]string{"NAME", "DESCRIPTION", "TEMPLATE"})
	triggers.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	triggers.SetCenterSeparator("|")
	triggers.SetAutoWrapText(false)
	for _, t := range builtin.Triggers {
		triggers.Append([]string{t.Name, t.Description, fmt.Sprintf("[%s](#%s)", t.Template, t.Template)})
	}
	triggers.Render()

	fmt.Println("")
	fmt.Println("## Templates")
	for _, t := range builtin.Templates {
		fmt.Fprintf(out, "### %s\n**title**: `%s`\n\n**body**:\n```\n%s\n```\n", t.Name, t.Title, t.Body)
	}
}

func main() {
	generate(os.Stdout)
}
