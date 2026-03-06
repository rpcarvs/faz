package cmd

import (
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
			cmd.Println("Warning: creating issue without description.")
			cmd.Println("  Issues without descriptions lack context for future work.")
			cmd.Println("  Consider adding --description \"Why this issue exists and what needs to be done\"")
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

		cmd.Printf("Created issue: %s\n", id)
		cmd.Printf("  Title: %s\n", args[0])
		cmd.Printf("  Type: %s\n", createType)
		cmd.Printf("  Priority: P%d\n", createPriority)
		cmd.Printf("  Status: open\n")
		if parentID != nil {
			cmd.Printf("  Parent: %s\n", *parentID)
		}
		return nil
	},
}

// init wires command flags and registration.
func init() {
	createCmd.Flags().StringVar(&createType, "type", "task", "Issue type (epic|task|bug|feature|chore|decision)")
	createCmd.Flags().IntVar(&createPriority, "priority", 2, "Issue priority (0-3)")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Issue description")
	createCmd.Flags().StringVar(&createParent, "parent", "", "Parent issue ID")
	rootCmd.AddCommand(createCmd)
}
