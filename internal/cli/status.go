package cli

import (
	"fmt"
	"os"
	"time"

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

			pr := repo.NewPersonRepo(db)
			or := repo.NewOrgRepo(db)
			dr := repo.NewDealRepo(db)
			tr := repo.NewTaskRepo(db)
			ir := repo.NewInteractionRepo(db)

			personCount, err := pr.Count(ctx)
			if err != nil {
				return err
			}

			orgCount, err := or.Count(ctx)
			if err != nil {
				return err
			}

			dealCount, dealValue, err := dr.OpenSummary(ctx)
			if err != nil {
				return err
			}

			overdueCount, err := tr.OverdueCount(ctx)
			if err != nil {
				return err
			}

			openTasks, err := tr.OpenCount(ctx)
			if err != nil {
				return err
			}

			sevenDaysAgo := time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02T15:04:05")
			weekInteractions, err := ir.CountSince(ctx, sevenDaysAgo)
			if err != nil {
				return err
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
			stages, err := dr.Pipeline(ctx)
			if err != nil {
				return err
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
				tasks, err := tr.FindAll(ctx, model.TaskFilters{Overdue: true, Limit: 5})
				if err != nil {
					return err
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
