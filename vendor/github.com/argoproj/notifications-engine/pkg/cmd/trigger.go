package cmd

import (
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/argoproj/notifications-engine/pkg/triggers"
	"github.com/argoproj/notifications-engine/pkg/util/misc"

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
		Use:   "run NAME RESOURCE_NAME",
		Short: "Evaluates specified trigger condition and prints the result",
		Example: fmt.Sprintf(`
# Execute trigger configured in 'argocd-notification-cm' ConfigMap
%s trigger run on-sync-status-unknown ./sample-app.yaml

# Execute trigger using my-config-map.yaml instead of '%s' ConfigMap
%s trigger run on-sync-status-unknown ./sample-app.yaml \
    --config-map ./my-config-map.yaml`, cmdContext.cliName, cmdContext.ConfigMapName, cmdContext.cliName),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("expected two arguments, got %d", len(args))
			}
			name := args[0]
			resourceName := args[1]
			api, err := cmdContext.getAPI()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to get api: %v\n", err)
				return nil
			}
			_, ok := api.GetConfig().Triggers[name]
			if !ok {
				var names []string
				for name := range api.GetConfig().Triggers {
					names = append(names, name)
				}
				_, _ = fmt.Fprintf(cmdContext.stderr,
					"trigger with name '%s' does not exist (found %s)\n", name, strings.Join(names, ", "))
				return nil
			}
			r, err := cmdContext.loadResource(resourceName)
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to load resource: %v\n", err)
				return nil
			}

			res, err := api.RunTrigger(name, r.Object)
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to execute trigger %s: %v\n", name, err)
				return nil
			}
			w := tabwriter.NewWriter(cmdContext.stdout, 5, 0, 2, ' ', 0)
			_, _ = fmt.Fprintf(w, "CONDITION\tRESULT\n")
			for i := range res {
				_, _ = fmt.Fprintf(w, "%s\t%v\n", api.GetConfig().Triggers[name][i].When, res[i].Triggered)
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
		Example: fmt.Sprintf(`
# prints all triggers
%s trigger get
# print YAML formatted on-sync-failed trigger definition
%s trigger get on-sync-failed -o=yaml
`, cmdContext.cliName, cmdContext.cliName),
		Short: "Prints information about configured triggers",
		RunE: func(c *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			items := map[string][]triggers.Condition{}

			api, err := cmdContext.getAPI()
			if err != nil {
				_, _ = fmt.Fprintf(cmdContext.stderr, "failed to get api: %v\n", err)
				return nil
			}
			for triggerName, trigger := range api.GetConfig().Triggers {
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
