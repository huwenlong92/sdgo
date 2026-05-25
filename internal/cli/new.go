package cli

import (
	"github.com/huwenlong/sdgo/internal/generator"
	"github.com/spf13/cobra"
)

func newNewCommand() *cobra.Command {
	var opt generator.ProjectOptions

	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Create a new project from the sdkitgo template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opt.Name = args[0]
			return generator.GenerateProject(".", opt)
		},
	}

	cmd.Flags().StringVar(&opt.ModulePath, "module", "", "Go module path")
	cmd.Flags().StringVar(&opt.SourceDir, "source", "", "sdkitgo source project path")
	cmd.Flags().BoolVar(&opt.Force, "force", false, "overwrite existing files")
	return cmd
}
