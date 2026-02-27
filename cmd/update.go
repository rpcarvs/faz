package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	updateTitle       string
	updateDescription string
	updateType        string
	updatePriority    int
	updateStatus      string
	updateParent      string
	clearParent       bool
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update issue fields",
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

		fields := map[string]any{}
		if cmd.Flags().Changed("title") {
			fields["title"] = updateTitle
		}
		if cmd.Flags().Changed("description") {
			fields["description"] = updateDescription
		}
		if cmd.Flags().Changed("type") {
			fields["type"] = updateType
		}
		if cmd.Flags().Changed("priority") {
			fields["priority"] = updatePriority
		}
		if cmd.Flags().Changed("status") {
			fields["status"] = updateStatus
		}
		if clearParent {
			fields["parent_public_id"] = (*string)(nil)
		} else if strings.TrimSpace(updateParent) != "" {
			normalizedParent, err := parseIDs([]string{updateParent})
			if err != nil {
				return err
			}
			fields["parent_public_id"] = &normalizedParent[0]
		}

		if err := svc.Update(ids[0], fields); err != nil {
			return err
		}

		fmt.Printf("Updated issue: %s\n", ids[0])
		fmt.Printf("  Status: updated\n")
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "Updated title")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "Updated description")
	updateCmd.Flags().StringVar(&updateType, "type", "", "Updated type")
	updateCmd.Flags().IntVar(&updatePriority, "priority", 2, "Updated priority")
	updateCmd.Flags().StringVar(&updateStatus, "status", "", "Updated status")
	updateCmd.Flags().StringVar(&updateParent, "parent", "", "Updated parent issue ID")
	updateCmd.Flags().BoolVar(&clearParent, "clear-parent", false, "Remove parent link")
	rootCmd.AddCommand(updateCmd)
}
