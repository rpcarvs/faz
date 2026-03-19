package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/rpcarvs/faz/internal/repo"
	"github.com/spf13/cobra"
)

var claimTTL time.Duration

var claimCmd = &cobra.Command{
	Use:   "claim <id>",
	Short: "Claim a work issue and move it to in_progress",
	Args:  cobra.ExactArgs(1),
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

		if err := svc.Claim(ids[0], claimTTL); err != nil {
			if errors.Is(err, repo.ErrIssueAlreadyClaimed) {
				return fmt.Errorf("this task is already claimed, try another one")
			}
			if errors.Is(err, repo.ErrIssueTypeNotClaimable) {
				return fmt.Errorf("this issue type cannot be claimed")
			}
			return err
		}

		cmd.Printf("Claimed issue: %s\n", ids[0])
		cmd.Printf("  Status: in_progress\n")
		cmd.Printf("  Lease TTL: %s\n", claimTTL)
		return nil
	},
}

// init wires command flags and registration.
func init() {
	claimCmd.Flags().DurationVar(&claimTTL, "ttl", 10*time.Minute, "Claim lease duration (example: 10m, 30m, 1h)")
	rootCmd.AddCommand(claimCmd)
}
