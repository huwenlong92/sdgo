package cli

import (
	"github.com/huwenlong/sdgo/internal/updater"
	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	var opt updater.Options

	cmd := &cobra.Command{
		Use:     "update [version]",
		Aliases: []string{"upgrade"},
		Short:   "Update the sdgo CLI.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opt.Version = args[0]
			}
			opt.Stdout = cmd.OutOrStdout()
			opt.Stderr = cmd.ErrOrStderr()
			return updater.Run(opt)
		},
	}

	cmd.Flags().StringVar(&opt.Target, "target", updater.DefaultInstallTarget, "go install target")
	return cmd
}
