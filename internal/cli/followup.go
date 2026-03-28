package cli

import (
	"fmt"
	"os"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/spf13/cobra"
)

func registerFollowupCommand(rootCmd *cobra.Command) {
	var days int

	cmd := &cobra.Command{
		Use:   "followup",
		Short: "Show overdue and upcoming follow-up tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			tr := repo.NewTaskRepo(db)

			// Overdue tasks
			overdue, err := tr.FindAll(cmd.Context(), model.TaskFilters{Overdue: true})
			if err != nil {
				return err
			}

			// Upcoming tasks (due within N days, default 3)
			upcoming, err := tr.FindAll(cmd.Context(), model.TaskFilters{DueWithinDays: &days})
			if err != nil {
				return err
			}

			f := resolveFormat()

			if len(overdue) > 0 {
				fmt.Fprintf(os.Stderr, "Overdue (%d):\n", len(overdue))
				if err := format.Output(os.Stdout, f, tasksToMaps(overdue), taskColumns, flagQuiet); err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr)
			}

			if len(upcoming) > 0 {
				fmt.Fprintf(os.Stderr, "Upcoming (%d, within %d days):\n", len(upcoming), days)
				if err := format.Output(os.Stdout, f, tasksToMaps(upcoming), taskColumns, flagQuiet); err != nil {
					return err
				}
			}

			if len(overdue) == 0 && len(upcoming) == 0 {
				fmt.Fprintln(os.Stderr, "No overdue or upcoming tasks.")
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 3, "number of days ahead to show upcoming tasks")
	rootCmd.AddCommand(cmd)
}
