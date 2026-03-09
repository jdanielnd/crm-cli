package cli

import (
	"fmt"
	"os"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/spf13/cobra"
)

func registerStatusCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Dashboard summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			ctx := cmd.Context()

			// Count people
			var personCount int
			if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM people WHERE archived = 0").Scan(&personCount); err != nil {
				return fmt.Errorf("count people: %w", err)
			}

			// Count organizations
			var orgCount int
			if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations WHERE archived = 0").Scan(&orgCount); err != nil {
				return fmt.Errorf("count organizations: %w", err)
			}

			// Open deals summary
			var dealCount int
			var dealValue float64
			if err := db.QueryRowContext(ctx,
				"SELECT COUNT(*), COALESCE(SUM(value), 0) FROM deals WHERE archived = 0 AND stage NOT IN ('won', 'lost')").
				Scan(&dealCount, &dealValue); err != nil {
				return fmt.Errorf("count deals: %w", err)
			}

			// Overdue tasks
			var overdueCount int
			if err := db.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM tasks WHERE archived = 0 AND completed = 0 AND due_at IS NOT NULL AND due_at < datetime('now')").
				Scan(&overdueCount); err != nil {
				return fmt.Errorf("count overdue tasks: %w", err)
			}

			// Interactions this week
			var weekInteractions int
			if err := db.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM interactions WHERE archived = 0 AND occurred_at >= datetime('now', '-7 days')").
				Scan(&weekInteractions); err != nil {
				return fmt.Errorf("count interactions: %w", err)
			}

			// Open tasks
			var openTasks int
			if err := db.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM tasks WHERE archived = 0 AND completed = 0").
				Scan(&openTasks); err != nil {
				return fmt.Errorf("count open tasks: %w", err)
			}

			f := resolveFormat()
			if f == format.FormatJSON || f == format.FormatCSV || f == format.FormatTSV {
				data := []map[string]any{{
					"contacts":           personCount,
					"organizations":      orgCount,
					"open_deals":         dealCount,
					"deal_value":         dealValue,
					"open_tasks":         openTasks,
					"overdue_tasks":      overdueCount,
					"interactions_7days": weekInteractions,
				}}
				cols := []format.ColumnDef{
					{Header: "Contacts", Field: "contacts"},
					{Header: "Organizations", Field: "organizations"},
					{Header: "Open Deals", Field: "open_deals"},
					{Header: "Deal Value", Field: "deal_value"},
					{Header: "Open Tasks", Field: "open_tasks"},
					{Header: "Overdue", Field: "overdue_tasks"},
					{Header: "Interactions (7d)", Field: "interactions_7days"},
				}
				return format.Output(os.Stdout, f, data, cols, flagQuiet)
			}

			// Human-readable table output
			fmt.Fprintf(os.Stdout, "%d contacts | %d organizations | %d open deals ($%.0f)\n",
				personCount, orgCount, dealCount, dealValue)
			fmt.Fprintf(os.Stdout, "%d open tasks | %d overdue | %d interactions this week\n",
				openTasks, overdueCount, weekInteractions)

			// Show pipeline if there are deals
			dr := repo.NewDealRepo(db)
			stages, err := dr.Pipeline(ctx)
			if err != nil {
				return fmt.Errorf("pipeline: %w", err)
			}
			if len(stages) > 0 {
				fmt.Fprintln(os.Stdout)
				fmt.Fprintln(os.Stdout, "Pipeline:")
				for _, s := range stages {
					fmt.Fprintf(os.Stdout, "  %-12s %d deals  $%.0f\n", s.Stage, s.Count, s.TotalValue)
				}
			}

			// Show overdue tasks if any
			if overdueCount > 0 {
				tr := repo.NewTaskRepo(db)
				tasks, err := tr.FindAll(ctx, model.TaskFilters{Overdue: true, Limit: 5})
				if err != nil {
					return fmt.Errorf("list overdue tasks: %w", err)
				}
				if len(tasks) > 0 {
					fmt.Fprintln(os.Stdout)
					fmt.Fprintln(os.Stdout, "Overdue tasks:")
					for _, t := range tasks {
						due := ""
						if t.DueAt != nil {
							due = *t.DueAt
						}
						fmt.Fprintf(os.Stdout, "  #%-4d %s (due %s)\n", t.ID, t.Title, due)
					}
				}
			}

			return nil
		},
	})
}
