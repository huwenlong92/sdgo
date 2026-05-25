package cli

import (
	"strings"

	"github.com/huwenlong/sdgo/internal/runner"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	var opt runner.Options

	cmd := &cobra.Command{
		Use:     "run [command...]",
		Aliases: []string{"dev"},
		Short:   "Run the current project with built-in hot reload.",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if looksLikeServeTarget(args) {
					opt.Target = args[0]
				} else {
					opt.Command = strings.Join(args, " ")
				}
			}
			return runner.Run(".", opt)
		},
	}

	cmd.Flags().StringVar(&opt.Command, "cmd", "", "command to run")
	cmd.Flags().StringVar(&opt.Watch, "watch", "", "comma-separated watch paths")
	cmd.Flags().BoolVar(&opt.NoWatch, "no-watch", false, "run without watching files")
	return cmd
}

func looksLikeServeTarget(args []string) bool {
	if len(args) != 1 {
		return false
	}
	arg := args[0]
	return !strings.ContainsAny(arg, " /\\.")
}
