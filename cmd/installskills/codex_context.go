package installskills

import (
	"fmt"
	"os"
	"path/filepath"

	"faz/internal/skillinstaller"
	"github.com/spf13/cobra"
)

// newCodexContextCommand installs or toggles the managed Codex context block.
func newCodexContextCommand() *cobra.Command {
	var local bool

	cmd := &cobra.Command{
		Use:   "codex-context",
		Short: "Install or toggle the faz task block in Codex AGENTS.md",
		Long: `Install or toggle the managed faz task-management block for Codex.

Default target: ~/.codex/AGENTS.md (or $CODEX_HOME/AGENTS.md).
Use --local to target ./AGENTS.md in the current directory.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetPath string
			var err error
			if local {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("resolve current directory: %w", err)
				}
				targetPath = filepath.Join(cwd, "AGENTS.md")
			} else {
				targetPath, err = skillinstaller.CodexContextPath()
				if err != nil {
					return err
				}
			}

			action, err := skillinstaller.InstallContextAtPath(targetPath)
			if err != nil {
				return err
			}

			if local {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated Codex local context (%s):\n", action)
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated Codex global context (%s):\n", action)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", targetPath)
			return nil
		},
	}
	cmd.Flags().BoolVar(&local, "local", false, "Write to ./AGENTS.md instead of global Codex context")
	return cmd
}
