package cmd

import (
	"fmt"
	"strings"

	"faz/internal/model"
	"github.com/spf13/cobra"
)

var (
	createType        string
	createPriority    int
	createDescription string
	createParent      string
)

var createCmd = &cobra.Command{
	Use:   "create \"title\"",
	Short: "Create an issue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		description := defaultDescription(createDescription)
		if description == "" {
			fmt.Println("Warning: creating issue without description.")
			fmt.Println("  Issues without descriptions lack context for future work.")
			fmt.Println("  Consider adding --description \"Why this issue exists and what needs to be done\"")
		}

		var parentID *string
		if strings.TrimSpace(createParent) != "" {
			normalizedParent, err := parseIDs([]string{createParent})
			if err != nil {
				return err
			}
			parentID = &normalizedParent[0]
		}

		id, err := svc.Create(model.Issue{
			Title:       args[0],
			Description: description,
			Type:        createType,
			Priority:    createPriority,
			Status:      "open",
			ParentID:    parentID,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Created issue: %s\n", id)
		fmt.Printf("  Title: %s\n", args[0])
		fmt.Printf("  Type: %s\n", createType)
		fmt.Printf("  Priority: P%d\n", createPriority)
		fmt.Printf("  Status: open\n")
		if parentID != nil {
			fmt.Printf("  Parent: %s\n", *parentID)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createType, "type", "task", "Issue type (epic|task|bug|feature|chore|decision)")
	createCmd.Flags().IntVar(&createPriority, "priority", 2, "Issue priority (0-3)")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Issue description")
	createCmd.Flags().StringVar(&createParent, "parent", "", "Parent issue ID")
	rootCmd.AddCommand(createCmd)
}
