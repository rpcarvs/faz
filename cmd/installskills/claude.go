package installskills

import (
	"fmt"

	"faz/internal/skillinstaller"
	"github.com/spf13/cobra"
)

func newClaudeSkillCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "claude-skill",
		Short: "Install the task-management-with-faz skill for Claude",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			installedPath, err := skillinstaller.InstallClaudeSkill(force)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Installed Claude skill:")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", installedPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing installed skill")
	return cmd
}
