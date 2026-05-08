package cmd

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/rpcarvs/faz/internal/db"
	"github.com/rpcarvs/faz/internal/tui/kanban"
	"github.com/spf13/cobra"
)

var kanbanPickEpic bool

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

		var opts []kanban.Option
		if kanbanPickEpic {
			opts = append(opts, kanban.WithPicker())
		}

		projectDir, err := currentProjectDir()
		if err != nil {
			return err
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := watcher.Close(); closeErr != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "close watcher: %v\n", closeErr)
			}
		}()
		if err := watcher.Add(filepath.Join(projectDir, db.DirName)); err != nil {
			return err
		}
		opts = append(opts, kanban.WithWatcher(watcher))

		model := kanban.NewModel(svc, opts...)
		program := tea.NewProgram(model, tea.WithAltScreen())
		_, err = program.Run()
		return err
	},
}

// init wires command flags and registration.
func init() {
	kanbanCmd.Flags().BoolVarP(&kanbanPickEpic, "epic", "e", false, "Open directly to the scope/epic picker")
	rootCmd.AddCommand(kanbanCmd)
}
