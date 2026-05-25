package cli

import (
	"github.com/huwenlong92/sdgo/internal/runner"
	"github.com/spf13/cobra"
)

func newServeCommand() *cobra.Command {
	var opt runner.Options

	cmd := &cobra.Command{
		Use:   "serve [target]",
		Short: "Run an sdkitgo serve target with hot reload.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opt.Target = args[0]
			}
			return runner.Run(".", opt)
		},
	}

	cmd.Flags().StringVar(&opt.Watch, "watch", "", "comma-separated watch paths")
	cmd.Flags().BoolVar(&opt.NoWatch, "no-watch", false, "run without watching files")
	return cmd
}
