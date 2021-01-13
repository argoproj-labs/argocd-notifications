package tools

import (
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/util/misc"
	"github.com/argoproj-labs/argocd-notifications/shared/legacy"
)

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
argocd-notifications template notify app-sync-succeeded guestbook --recipient slack:argocd-notifications

# Render notification render generated notification in console
argocd-notifications template notify app-sync-succeeded guestbook
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
			config.API.AddNotificationService("console", services.NewConsoleService(cmdContext.stdout))

			app, err := cmdContext.loadApplication(application)
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to load application: %v\n", err)
				return nil
			}

			for _, recipient := range recipients {
				parts := strings.Split(recipient, ":")
				dest := services.Destination{Service: parts[0]}
				if len(parts) > 1 {
					dest.Recipient = parts[1]
				}
				vars := map[string]interface{}{"app": app.Object, "context": legacy.InjectLegacyVar(config.Context, dest.Service)}
				if err := config.API.Send(vars, []string{name}, dest); err != nil {
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
argocd-notifications template get
# print YAML formatted app-sync-succeeded template definition
argocd-notifications template get app-sync-succeeded -o=yaml
`,
		Short: "Prints information about configured templates",
		RunE: func(c *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			items := map[string]services.Notification{}

			config, err := cmdContext.getConfig()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to parse config: %v\n", err)
				return nil
			}
			for n, template := range config.Templates {
				if n == name || name == "" {
					items[n] = template
				}
			}
			switch output {
			case "", "wide":
				w := tabwriter.NewWriter(cmdContext.stdout, 5, 0, 2, ' ', 0)
				_, _ = fmt.Fprintf(w, "NAME\tPREVIEW\n")
				misc.IterateStringKeyMap(items, func(name string) {
					template := items[name]
					_, _ = fmt.Fprintf(w, "%s\t%s\n", name, template.Preview())
				})
				_ = w.Flush()
			case "name":
				misc.IterateStringKeyMap(items, func(name string) {
					_, _ = fmt.Println(name)
				})
			default:
				return misc.PrintFormatted(items, output, cmdContext.stdout)
			}
			return nil
		},
	}
	addOutputFlags(&command, &output)
	return &command
}
