package cmd

import (
	"fmt"
	"time"

	"github.com/rpcarvs/faz/internal/model"
	"github.com/spf13/cobra"
)

var (
	monitorIntervalSeconds int
	monitorAll             bool
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Continuously refresh task list output",
	Long:  "Monitor loops over list output with a terminal refresh and a configurable interval.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if monitorIntervalSeconds <= 0 {
			return fmt.Errorf("interval must be greater than 0")
		}

		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		filter := model.ListFilter{All: monitorAll}
		interval := time.Duration(monitorIntervalSeconds) * time.Second

		for {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), "\033[H\033[2J")

			issues, err := svc.List(filter)
			if err != nil {
				return err
			}
			printIssueList(cmd.OutOrStdout(), issues)

			time.Sleep(interval)
		}
	},
}

// init wires command flags and registration.
func init() {
	monitorCmd.Flags().IntVarP(&monitorIntervalSeconds, "interval", "t", 5, "Refresh interval in seconds")
	monitorCmd.Flags().BoolVar(&monitorAll, "all", false, "Include closed issues")
	rootCmd.AddCommand(monitorCmd)
}
