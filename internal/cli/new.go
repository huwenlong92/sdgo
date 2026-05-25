package cli

import (
	"github.com/huwenlong92/sdgo/internal/generator"
	"github.com/spf13/cobra"
)

func newNewCommand() *cobra.Command {
	var opt generator.ProjectOptions

	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Create a new project from a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opt.Name = args[0]
			return generator.GenerateProject(".", opt)
		},
	}

	cmd.Flags().StringVar(&opt.ModulePath, "module", "", "Go module path for Go templates")
	cmd.Flags().StringVar(&opt.Template, "template", "", "template name used for default source lookup")
	cmd.Flags().StringVar(&opt.SourceDir, "source", "", "template source project path or Git URL")
	cmd.Flags().StringVar(&opt.Branch, "branch", "", "Git branch or tag to clone when --source is a Git URL")
	cmd.Flags().BoolVar(&opt.Force, "force", false, "overwrite existing files")
	return cmd
}
