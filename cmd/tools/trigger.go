package tools

import (
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/argoproj-labs/argocd-notifications/triggers"

	"github.com/spf13/cobra"
)

func newTriggerCommand(cmdContext *commandContext) *cobra.Command {
	var command = cobra.Command{
		Use:   "trigger",
		Short: "Notification triggers related commands",
		RunE: func(c *cobra.Command, args []string) error {
			return errors.New("select child command")
		},
	}
	command.AddCommand(newTriggerRunCommand(cmdContext))
	command.AddCommand(newTriggerGetCommand(cmdContext))

	return &command
}

func newTriggerRunCommand(cmdContext *commandContext) *cobra.Command {
	var command = cobra.Command{
		Use:   "run NAME APPLICATION",
		Short: "Evaluates specified trigger condition and prints the result",
		Example: `
# Execute trigger configured in 'argocd-notification-cm' ConfigMap
argocd-notifications tools trigger run on-sync-status-unknown ./sample-app.yaml

# Execute trigger using argocd-notifications-cm.yaml instead of 'argocd-notification-cm' ConfigMap
argocd-notifications tools trigger run on-sync-status-unknown ./sample-app.yaml \
    --config-map ./argocd-notifications-cm.yaml`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("expected two arguments, got %d", len(args))
			}
			name := args[0]
			application := args[1]
			cfg, err := cmdContext.getConfig()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to parse config: %v\n", err)
				return nil
			}
			trigger, ok := cfg.Triggers[name]
			if !ok {
				var names []string
				for name := range cfg.Triggers {
					names = append(names, name)
				}
				_, _ = fmt.Fprintf(cmdContext.stderr,
					"trigger with name '%s' does not exist (found %s)\n", name, strings.Join(names, ", "))
				return nil
			}
			app, err := cmdContext.loadApplication(application)
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to load application: %v\n", err)
				return nil
			}
			ok, err = trigger.Triggered(app)
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to execute trigger %s: %v\n", name, err)
				return nil
			}
			_, _ = fmt.Fprintf(cmdContext.stdout, "%v\n", ok)
			return nil
		},
	}

	return &command
}

func newTriggerGetCommand(cmdContext *commandContext) *cobra.Command {
	var (
		output string
	)
	var command = cobra.Command{
		Use: "get",
		Example: `
# prints all triggers
argocd-notifications tools trigger get
# print YAML formatted on-sync-failed trigger definition
argocd-notifications tools trigger get on-sync-failed -o=yaml
`,
		Short: "Prints information about configured triggers",
		RunE: func(c *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			var items []triggers.NotificationTrigger

			config, err := cmdContext.getConfig()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to parse config: %v\n", err)
				return nil
			}
			for _, trigger := range config.TriggersSettings {
				if trigger.Name == name || name == "" {
					items = append(items, trigger)
				}
			}
			switch output {
			case "", "wide":
				w := tabwriter.NewWriter(cmdContext.stdout, 5, 0, 2, ' ', 0)
				_, _ = fmt.Fprintf(w, "NAME\tENABLED\tTEMPLATE\tCONDITION\n")
				for _, trigger := range items {
					enabled := "true"
					if trigger.Enabled != nil && !*trigger.Enabled {
						enabled = "false"
					}
					_, _ = fmt.Fprintf(w, "%s\t%v\t%s\t%s\n",
						trigger.Name, enabled, trigger.Template, trigger.Condition)
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
