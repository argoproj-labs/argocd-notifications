package tools

import (
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/argoproj-labs/argocd-notifications/expr"
	"github.com/argoproj-labs/argocd-notifications/pkg/triggers"
	"github.com/argoproj-labs/argocd-notifications/pkg/util/misc"

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
argocd-notifications trigger run on-sync-status-unknown ./sample-app.yaml

# Execute trigger using argocd-notifications-cm.yaml instead of 'argocd-notification-cm' ConfigMap
argocd-notifications trigger run on-sync-status-unknown ./sample-app.yaml \
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
			_, ok := cfg.Triggers[name]
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
			res, err := cfg.API.RunTrigger(name, expr.Spawn(app, cfg.ArgoCDService, map[string]interface{}{"app": app.Object}))
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to execute trigger %s: %v\n", name, err)
				return nil
			}
			w := tabwriter.NewWriter(cmdContext.stdout, 5, 0, 2, ' ', 0)
			_, _ = fmt.Fprintf(w, "CONDITION\tRESULT\n")
			for i := range res {
				_, _ = fmt.Fprintf(w, "%s\t%v\n", cfg.Triggers[name][i].When, res[i].Triggered)
			}
			_ = w.Flush()
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
argocd-notifications trigger get
# print YAML formatted on-sync-failed trigger definition
argocd-notifications trigger get on-sync-failed -o=yaml
`,
		Short: "Prints information about configured triggers",
		RunE: func(c *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			items := map[string][]triggers.Condition{}

			config, err := cmdContext.getConfig()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to parse config: %v\n", err)
				return nil
			}
			for triggerName, trigger := range config.Triggers {
				if triggerName == name || name == "" {
					items[triggerName] = trigger
				}
			}
			switch output {
			case "", "wide":
				w := tabwriter.NewWriter(cmdContext.stdout, 5, 0, 2, ' ', 0)
				_, _ = fmt.Fprintf(w, "NAME\tTEMPLATE\tCONDITION\n")
				misc.IterateStringKeyMap(items, func(triggerName string) {
					for i, condition := range items[triggerName] {
						name := triggerName
						if i > 0 {
							name = ""
						}
						_, _ = fmt.Fprintf(w, "%s\t%v\t%s\n", name, strings.Join(condition.Send, ", "), condition.When)
					}
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
