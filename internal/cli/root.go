package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

const Version = "0.1.1"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sdgo",
		Short: "sdgo creates and runs sdkitgo projects.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newNewCommand())
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newServeCommand())
	cmd.AddCommand(newCompletionCommand())
	cmd.AddCommand(newUpgradeCommand())
	cmd.AddCommand(newTemplateCommand())
	return cmd
}

func versionString() string {
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return Version
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the sdgo CLI version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), versionString())
		},
	}
}
