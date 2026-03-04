package installskills

import "github.com/spf13/cobra"

// NewCommand builds the install parent command and registers skill installers.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install built-in agent skills and context blocks",
	}

	cmd.AddCommand(newCodexSkillCommand())
	cmd.AddCommand(newClaudeSkillCommand())
	cmd.AddCommand(newCodexContextCommand())
	cmd.AddCommand(newClaudeContextCommand())
	return cmd
}
