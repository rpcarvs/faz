package installskills

import (
	"fmt"

	"faz/internal/skillinstaller"
	"github.com/spf13/cobra"
)

// newCodexSkillCommand installs the bundled Codex skill package.
func newCodexSkillCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "codex-skill",
		Short: "Install the task-management-with-faz skill for Codex",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			installedPath, err := skillinstaller.InstallCodexSkill(force)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Installed Codex skill:")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", installedPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing installed skill")
	return cmd
}
