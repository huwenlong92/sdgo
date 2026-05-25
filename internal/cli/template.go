package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/huwenlong92/sdgo/internal/generator"
	"github.com/spf13/cobra"
)

func newTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"templates"},
		Short:   "Show available project templates.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateList(cmd)
		},
	}

	cmd.AddCommand(newTemplateListCommand())
	cmd.AddCommand(newTemplateInfoCommand())
	return cmd
}

func newTemplateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available project templates.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateList(cmd)
		},
	}
}

func newTemplateInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info <template>",
		Short: "Show project template details.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			info, ok := generator.BuiltinTemplate(args[0])
			if !ok {
				return fmt.Errorf("template not found: %s", args[0])
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Name: %s\n", info.Name)
			fmt.Fprintf(out, "Kind: %s\n", info.Kind)
			fmt.Fprintf(out, "Default: %t\n", info.Default)
			fmt.Fprintf(out, "Environment: %s\n", generator.TemplateSourceEnv(info.Name))
			fmt.Fprintf(out, "Description: %s\n", info.Description)
			fmt.Fprintln(out, "Sources:")
			for _, source := range info.Sources {
				fmt.Fprintf(out, "  - %s\n", source)
			}
			return nil
		},
	}
}

func runTemplateList(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tKIND\tDEFAULT\tSOURCE")
	for _, info := range generator.BuiltinTemplates() {
		fmt.Fprintf(w, "%s\t%s\t%t\t%s\n", info.Name, info.Kind, info.Default, strings.Join(info.Sources, ","))
	}
	return w.Flush()
}
