package cmd

import (
	"errors"
	"fmt"
	"time"

	"faz/internal/repo"
	"github.com/spf13/cobra"
)

var claimTTL time.Duration

var claimCmd = &cobra.Command{
	Use:   "claim <id>",
	Short: "Claim an issue for work and move it to in_progress",
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
				fmt.Println("This task is already claimed, try another one")
				return nil
			}
			return err
		}

		fmt.Printf("Claimed issue: %s\n", ids[0])
		fmt.Printf("  Status: in_progress\n")
		fmt.Printf("  Lease TTL: %s\n", claimTTL)
		return nil
	},
}

func init() {
	claimCmd.Flags().DurationVar(&claimTTL, "ttl", time.Hour, "Claim lease duration (example: 30m, 1h)")
	rootCmd.AddCommand(claimCmd)
}
