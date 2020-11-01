package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/argoproj-labs/argocd-notifications/cmd/tools"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/spf13/cobra/doc"
)

func generateBuiltInTriggersDocs(out io.Writer, builtin *settings.Config) {
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

func generateCommandsDocs(out io.Writer) error {
	toolsCmd := tools.NewToolsCommand()
	for _, subCommand := range toolsCmd.Commands() {
		for _, cmd := range subCommand.Commands() {
			var cmdDesc bytes.Buffer
			if err := doc.GenMarkdown(cmd, &cmdDesc); err != nil {
				return err
			}
			for _, line := range strings.Split(cmdDesc.String(), "\n") {
				if strings.HasPrefix(line, "### SEE ALSO") {
					break
				}
				_, _ = fmt.Fprintf(out, "%s\n", line)
			}
		}
	}
	return nil
}

func main() {
	var builtItDocsData bytes.Buffer
	wd, err := os.Getwd()
	dieOnError(err, "Failed to get current working directory")

	templatesDir := path.Join(wd, "builtin/templates")
	triggersDir := path.Join(wd, "builtin/triggers")

	cnf, err := tools.BuildConfigFromFS(templatesDir, triggersDir)
	dieOnError(err, "Failed to build builtin config")
	generateBuiltInTriggersDocs(&builtItDocsData, cnf)
	if err := ioutil.WriteFile("./docs/built-in.md", builtItDocsData.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
	var commandDocs bytes.Buffer
	if err := generateCommandsDocs(&commandDocs); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("./docs/troubleshooting-commands.md", commandDocs.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

func dieOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("[ERROR] %s: %v", msg, err)
		os.Exit(1)
	}
}
