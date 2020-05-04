package tools

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/argoproj-labs/argocd-notifications/notifiers"

	"github.com/spf13/cobra"

	sharedrecipients "github.com/argoproj-labs/argocd-notifications/shared/recipients"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

type consoleNotifier struct {
	stdout io.Writer
}

func (c *consoleNotifier) Send(notification notifiers.Notification, _ string) error {
	return printFormatted(notification, "yaml", c.stdout)
}

func newTemplateCommand(cmdContext *commandContext) *cobra.Command {
	var command = cobra.Command{
		Use: "template",
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
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("expected two arguments, got %d", len(args))
			}
			name := args[0]
			application := args[1]

			_, notifiersByName, config, err := cmdContext.getConfig()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to parse config: %v\n", err)
				return nil
			}
			notifiersByName["console"] = &consoleNotifier{stdout: cmdContext.stdout}
			triggersByName, err := triggers.GetTriggers(config.Templates, []triggers.NotificationTrigger{{
				Name:      "__test__",
				Template:  name,
				Condition: "true",
			}})
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to parse config: %v\n", err)
				return nil
			}
			trigger := triggersByName["__test__"]

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
				notifierType := parts[0]
				notifier, ok := notifiersByName[notifierType]
				if !ok {
					_, _ = fmt.Fprintf(cmdContext.stderr, "%s is not valid recipient type.", notifierType)
					return nil
				}

				ctx := sharedrecipients.CopyStringMap(config.Context)
				ctx["notificationType"] = notifierType
				notification, err := trigger.FormatNotification(app, ctx)
				if err != nil {
					_, _ = fmt.Fprintf(cmdContext.stderr, "failed to format notification: %v", err)
				}
				if err = notifier.Send(*notification, parts[1]); err != nil {
					_, _ = fmt.Fprintf(cmdContext.stderr, "failed to notify '%s': %v", recipient, err)
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
		RunE: func(c *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			var items []triggers.NotificationTemplate

			_, _, config, err := cmdContext.getConfig()
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
