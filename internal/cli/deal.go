package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/spf13/cobra"
)

var dealColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Title", Field: "title"},
	{Header: "Value", Field: "value"},
	{Header: "Stage", Field: "stage"},
	{Header: "Person", Field: "person_id"},
	{Header: "Org", Field: "org_id"},
}

var pipelineColumns = []format.ColumnDef{
	{Header: "Stage", Field: "stage"},
	{Header: "Count", Field: "count"},
	{Header: "Total Value", Field: "total_value"},
}

func dealToMap(d *model.Deal) map[string]any {
	m := map[string]any{
		"id":         d.ID,
		"uuid":       d.UUID,
		"title":      d.Title,
		"stage":      d.Stage,
		"created_at": d.CreatedAt,
		"updated_at": d.UpdatedAt,
	}
	if d.Value != nil {
		m["value"] = *d.Value
	}
	if d.PersonID != nil {
		m["person_id"] = *d.PersonID
	}
	if d.OrgID != nil {
		m["org_id"] = *d.OrgID
	}
	if d.Notes != nil {
		m["notes"] = *d.Notes
	}
	if d.ClosedAt != nil {
		m["closed_at"] = *d.ClosedAt
	}
	return m
}

func dealsToMaps(deals []*model.Deal) []map[string]any {
	result := make([]map[string]any, len(deals))
	for i, d := range deals {
		result[i] = dealToMap(d)
	}
	return result
}

func registerDealCommands(rootCmd *cobra.Command) {
	dealCmd := &cobra.Command{
		Use:   "deal",
		Short: "Manage deals and pipeline",
	}

	dealCmd.AddCommand(dealAddCmd())
	dealCmd.AddCommand(dealListCmd())
	dealCmd.AddCommand(dealShowCmd())
	dealCmd.AddCommand(dealEditCmd())
	dealCmd.AddCommand(dealDeleteCmd())
	dealCmd.AddCommand(dealPipelineCmd())

	rootCmd.AddCommand(dealCmd)
}

func dealAddCmd() *cobra.Command {
	var value float64
	var stage, notes string
	var personID, orgID int64

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new deal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			if !model.ValidDealStage(stage) {
				return model.NewExitError(model.ErrValidation, "invalid stage: %s (must be one of: %s)", stage, strings.Join(model.DealStages, ", "))
			}

			input := model.CreateDealInput{
				Title: args[0],
				Stage: stage,
				Notes: nilIfEmpty(notes),
			}
			if cmd.Flags().Changed("value") {
				input.Value = &value
			}
			if cmd.Flags().Changed("person") {
				input.PersonID = &personID
			}
			if cmd.Flags().Changed("org") {
				input.OrgID = &orgID
			}

			r := repo.NewDealRepo(db)
			deal, err := r.Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			data := []map[string]any{dealToMap(deal)}
			return format.Output(os.Stdout, resolveFormat(), data, dealColumns, flagQuiet)
		},
	}

	cmd.Flags().Float64Var(&value, "value", 0, "deal value")
	cmd.Flags().StringVar(&stage, "stage", "lead", "deal stage: "+strings.Join(model.DealStages, ", "))
	cmd.Flags().StringVar(&notes, "notes", "", "notes")
	cmd.Flags().Int64Var(&personID, "person", 0, "associated person ID")
	cmd.Flags().Int64Var(&orgID, "org", 0, "associated organization ID")

	return cmd
}

func dealListCmd() *cobra.Command {
	var stage string
	var personID, orgID int64
	var limit int
	var open bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deals",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			filters := model.DealFilters{Limit: limit, ExcludeClosed: open}
			if stage != "" {
				filters.Stage = &stage
			}
			if cmd.Flags().Changed("person") {
				filters.PersonID = &personID
			}
			if cmd.Flags().Changed("org") {
				filters.OrgID = &orgID
			}

			r := repo.NewDealRepo(db)
			deals, err := r.FindAll(cmd.Context(), filters)
			if err != nil {
				return err
			}

			return format.Output(os.Stdout, resolveFormat(), dealsToMaps(deals), dealColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&stage, "stage", "", "filter by stage")
	cmd.Flags().Int64Var(&personID, "person", 0, "filter by person ID")
	cmd.Flags().Int64Var(&orgID, "org", 0, "filter by organization ID")
	cmd.Flags().IntVar(&limit, "limit", 0, "max results")
	cmd.Flags().BoolVar(&open, "open", false, "show only open deals (exclude won/lost)")

	return cmd
}

func dealShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show deal details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid deal ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewDealRepo(db)
			deal, err := r.FindByID(cmd.Context(), id)
			if err != nil {
				return err
			}

			data := []map[string]any{dealToMap(deal)}
			return format.Output(os.Stdout, resolveFormat(), data, dealColumns, flagQuiet)
		},
	}
}

func dealEditCmd() *cobra.Command {
	var title, stage, notes, closedAt string
	var value float64
	var personID, orgID int64

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a deal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid deal ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			input := model.UpdateDealInput{}
			if cmd.Flags().Changed("title") {
				input.Title = &title
			}
			if cmd.Flags().Changed("value") {
				input.Value = &value
			}
			if cmd.Flags().Changed("stage") {
				if !model.ValidDealStage(stage) {
					return model.NewExitError(model.ErrValidation, "invalid stage: %s (must be one of: %s)", stage, strings.Join(model.DealStages, ", "))
				}
				input.Stage = &stage
			}
			if cmd.Flags().Changed("person") {
				input.PersonID = &personID
			}
			if cmd.Flags().Changed("org") {
				input.OrgID = &orgID
			}
			if cmd.Flags().Changed("notes") {
				input.Notes = &notes
			}
			if cmd.Flags().Changed("closed-at") {
				input.ClosedAt = &closedAt
			}

			r := repo.NewDealRepo(db)
			deal, err := r.Update(cmd.Context(), id, input)
			if err != nil {
				return err
			}

			data := []map[string]any{dealToMap(deal)}
			return format.Output(os.Stdout, resolveFormat(), data, dealColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "deal title")
	cmd.Flags().Float64Var(&value, "value", 0, "deal value")
	cmd.Flags().StringVar(&stage, "stage", "", "deal stage: "+strings.Join(model.DealStages, ", "))
	cmd.Flags().StringVar(&notes, "notes", "", "notes")
	cmd.Flags().Int64Var(&personID, "person", 0, "associated person ID")
	cmd.Flags().Int64Var(&orgID, "org", 0, "associated organization ID")
	cmd.Flags().StringVar(&closedAt, "closed-at", "", "date the deal was closed")

	return cmd
}

func dealDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a deal (soft-delete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid deal ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewDealRepo(db)
			if err := r.Archive(cmd.Context(), id); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Deleted deal #%d\n", id)
			return nil
		},
	}
}

func dealPipelineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pipeline",
		Short: "Show deal pipeline summary by stage",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewDealRepo(db)
			stages, err := r.Pipeline(cmd.Context())
			if err != nil {
				return err
			}

			data := make([]map[string]any, len(stages))
			for i, s := range stages {
				data[i] = map[string]any{
					"stage":       s.Stage,
					"count":       s.Count,
					"total_value": s.TotalValue,
				}
			}
			return format.Output(os.Stdout, resolveFormat(), data, pipelineColumns, flagQuiet)
		},
	}
}
