package installskills

import (
	"fmt"
	"os"
	"path/filepath"

	"faz/internal/skillinstaller"
	"github.com/spf13/cobra"
)

// newClaudeContextCommand installs or updates the managed Claude context block.
func newClaudeContextCommand() *cobra.Command {
	var local bool

	cmd := &cobra.Command{
		Use:   "claude-context",
		Short: "Install or update the faz task block in Claude CLAUDE.md",
		Long: `Install or update the managed faz task-management block for Claude.

Default target: ~/.claude/CLAUDE.md.
Use --local to target ./CLAUDE.md in the current directory.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetPath string
			var err error
			if local {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("resolve current directory: %w", err)
				}
				targetPath = filepath.Join(cwd, "CLAUDE.md")
			} else {
				targetPath, err = skillinstaller.ClaudeContextPath()
				if err != nil {
					return err
				}
			}

			action, err := skillinstaller.InstallContextAtPath(targetPath)
			if err != nil {
				return err
			}

			if local {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated Claude local context (%s):\n", action)
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated Claude global context (%s):\n", action)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", targetPath)
			return nil
		},
	}
	cmd.Flags().BoolVar(&local, "local", false, "Write to ./CLAUDE.md instead of global Claude context")
	return cmd
}
