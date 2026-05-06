package installskills

import "github.com/spf13/cobra"

// ProjectRootFunc resolves the current Git repository root for local installs.
type ProjectRootFunc func() (string, error)

// NewCommand builds the install parent command and registers skill installers.
func NewCommand(projectRoot ProjectRootFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install built-in agent integration for Codex or Claude",
	}

	cmd.AddCommand(newProviderCommand("codex", projectRoot))
	cmd.AddCommand(newProviderCommand("claude", projectRoot))
	return cmd
}
