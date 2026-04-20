package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rpcarvs/faz/internal/tui/kanban"
	"github.com/spf13/cobra"
)

var kanbanCmd = &cobra.Command{
	Use:   "kanban",
	Short: "Open a read-only kanban TUI for tasks",
	Long:  "Kanban opens a read-only terminal view of faz tasks grouped into TO DO, CLAIMED, and DONE columns across all epics or a selected epic scope.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		model := kanban.NewModel(svc)
		program := tea.NewProgram(model, tea.WithAltScreen())
		_, err = program.Run()
		return err
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(kanbanCmd)
}
