package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sdgo",
		Short: "sdgo creates and runs sdkitgo projects.",
	}

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newNewCommand())
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newServeCommand())
	cmd.AddCommand(newGenCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newTemplateCommand())
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the sdgo CLI version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), Version)
		},
	}
}
