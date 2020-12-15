package tools

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/argoproj-labs/argocd-notifications/pkg/templates"

	"github.com/spf13/cobra"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

type consoleService struct {
	stdout io.Writer
}

func (c *consoleService) Send(notification services.Notification, _ string) error {
	return printFormatted(notification, "yaml", c.stdout)
}

func newTemplateCommand(cmdContext *commandContext) *cobra.Command {
	var command = cobra.Command{
		Use:   "template",
		Short: "Notification templates related commands",
		RunE: func(c *cobra.Command, args []string) error {
			return errors.New("select child command")
		},
	}
	command.AddCommand(newTemplateNotifyCommand(cmdContext))
	command.AddCommand(newTemplateGetCommand(cmdContext))

	return &command
}
func newTemplateNotifyCommand(cmdContext *commandContext) *cobra.Command {
	var (
		recipients []string
	)
	var command = cobra.Command{
		Use: "notify NAME APPLICATION",
		Example: `
# Trigger notification using in-cluster config map and secret
argocd-notifications tools template notify app-sync-succeeded guestbook --recipient slack:argocd-notifications

# Render notification render generated notification in console
argocd-notifications tools template notify app-sync-succeeded guestbook
`,
		Short: "Generates notification using the specified template and send it to specified recipients",
		RunE: func(c *cobra.Command, args []string) error {
			cancel := withDebugLogs()
			defer cancel()
			if len(args) < 2 {
				return fmt.Errorf("expected two arguments, got %d", len(args))
			}
			name := args[0]
			application := args[1]

			config, err := cmdContext.getConfig()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to parse config: %v\n", err)
				return nil
			}
			config.Notifier.AddService("console", &consoleService{stdout: cmdContext.stdout})

			app, err := cmdContext.loadApplication(application)
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to load application: %v\n", err)
				return nil
			}

			for _, recipient := range recipients {
				parts := strings.Split(recipient, ":")
				if len(parts) < 2 {
					_, _ = fmt.Fprintf(cmdContext.stderr, "%s is not valid recipient. Expected recipient format is <type>:<name>\n", recipient)
					return nil
				}
				serviceType := parts[0]

				vars := map[string]interface{}{"app": app.Object, "context": config.Context}
				if err := config.Notifier.Send(vars, name, serviceType, parts[1]); err != nil {
					_, _ = fmt.Fprintf(cmdContext.stderr, "failed to notify '%s': %v\n", recipient, err)
					return nil
				}
			}

			return nil
		},
	}
	command.Flags().StringArrayVar(&recipients, "recipient", []string{"console:stdout"}, "List of recipients")

	return &command
}

func newTemplateGetCommand(cmdContext *commandContext) *cobra.Command {
	var (
		output string
	)
	var command = cobra.Command{
		Use: "get",
		Example: `
# prints all templates
argocd-notifications tools template get
# print YAML formatted app-sync-succeeded template definition
argocd-notifications tools template get app-sync-succeeded -o=yaml
`,
		Short: "Prints information about configured templates",
		RunE: func(c *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			var items []templates.NotificationTemplate

			config, err := cmdContext.getConfig()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to parse config: %v\n", err)
				return nil
			}
			for _, template := range config.Templates {
				if template.Name == name || name == "" {
					items = append(items, template)
				}
			}
			switch output {
			case "", "wide":
				w := tabwriter.NewWriter(cmdContext.stdout, 5, 0, 2, ' ', 0)
				_, _ = fmt.Fprintf(w, "NAME\tTITLE\n")
				for _, template := range items {
					_, _ = fmt.Fprintf(w, "%s\t%s\n", template.Name, template.Title)
				}
				_ = w.Flush()
			case "name":
				for i := range items {
					_, _ = fmt.Println(items[i].Name)
				}
			default:
				return printFormatted(items, output, cmdContext.stdout)
			}
			return nil
		},
	}
	addOutputFlags(&command, &output)
	return &command
}
