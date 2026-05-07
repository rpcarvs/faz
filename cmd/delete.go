package cmd

import "github.com/spf13/cobra"

var deleteCmd = &cobra.Command{
	Use:   "delete <id> [id...]",
	Short: "Permanently delete one or more issues",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		ids, err := parseIDs(args)
		if err != nil {
			return err
		}

		for _, id := range ids {
			if err := svc.Delete(id); err != nil {
				return err
			}
			stdoutPrintf(cmd, "Deleted issue: %s\n", id)
		}
		return nil
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(deleteCmd)
}
