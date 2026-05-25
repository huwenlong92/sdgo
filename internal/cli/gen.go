package cli

import (
	"github.com/huwenlong/sdgo/internal/generator"
	"github.com/spf13/cobra"
)

func newGenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate project code.",
	}
	cmd.AddCommand(newGenModuleCommand())
	return cmd
}

func newGenModuleCommand() *cobra.Command {
	var opt generator.ModuleOptions

	cmd := &cobra.Command{
		Use:   "module <name>",
		Short: "Generate a module skeleton.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opt.Name = args[0]
			return generator.GenerateModule(".", opt)
		},
	}

	cmd.Flags().BoolVar(&opt.WithDocs, "with-docs", true, "generate module documentation")
	cmd.Flags().BoolVar(&opt.Force, "force", false, "overwrite existing files")
	return cmd
}
