package installskills

import (
	"fmt"
	"strings"

	"github.com/rpcarvs/faz/internal/skillinstaller"
	"github.com/spf13/cobra"
)

// newProviderCommand installs the full faz integration for one agent provider.
func newProviderCommand(name string, projectRoot ProjectRootFunc) *cobra.Command {
	var local bool
	var force bool

	cmd := &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Install faz integration for %s", providerLabel(name)),
		Long: fmt.Sprintf(`Install the faz integration for %[1]s.

This installs the task-management-with-faz skill, the managed task context block,
and a SessionStart hook that runs faz init and faz onboard.

Use --local to install into the current Git repository instead of the global
%[1]s configuration.`, providerLabel(name)),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			options := skillinstaller.InstallOptions{
				Provider: skillinstaller.Provider(name),
				Local:    local,
				Force:    force,
			}
			if local {
				if projectRoot == nil {
					return fmt.Errorf("local install requires a project root resolver")
				}
				root, err := projectRoot()
				if err != nil {
					return err
				}
				options.LocalRoot = root
			}

			result, err := skillinstaller.InstallProvider(options)
			if err != nil {
				return err
			}
			printInstallResult(cmd, name, local, result)
			return nil
		},
	}

	cmd.Flags().BoolVar(&local, "local", false, "Install into the current Git repository")
	cmd.Flags().BoolVar(&force, "force", false, "Replace existing skill directory before installing")
	return cmd
}

// printInstallResult writes a concise summary of installed integration files.
func printInstallResult(cmd *cobra.Command, provider string, local bool, result skillinstaller.InstallResult) {
	scope := "global"
	if local {
		scope = "local"
	}
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Installed %s %s integration:\n", providerLabel(provider), scope)
	_, _ = fmt.Fprintf(out, "  Skill: %s\n", result.SkillPath)
	_, _ = fmt.Fprintf(out, "  Context (%s): %s\n", result.ContextAction, result.ContextPath)
	_, _ = fmt.Fprintf(out, "  Hook (%s): %s\n", result.HookAction, result.HookPath)
	if result.ClaudePointerPath != "" {
		_, _ = fmt.Fprintf(out, "  Claude pointer (%s): %s\n", result.ClaudePointerAction, result.ClaudePointerPath)
	}
}

// providerLabel returns a user-facing provider name.
func providerLabel(provider string) string {
	if provider == "" {
		return ""
	}
	return strings.ToUpper(provider[:1]) + provider[1:]
}
