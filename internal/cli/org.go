package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/spf13/cobra"
)

var orgColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Name", Field: "name"},
	{Header: "Domain", Field: "domain"},
	{Header: "Industry", Field: "industry"},
}

func orgToMap(o *model.Organization) map[string]any {
	m := map[string]any{
		"id":         o.ID,
		"uuid":       o.UUID,
		"name":       o.Name,
		"created_at": o.CreatedAt,
		"updated_at": o.UpdatedAt,
	}
	if o.Domain != nil {
		m["domain"] = *o.Domain
	}
	if o.Industry != nil {
		m["industry"] = *o.Industry
	}
	if o.Notes != nil {
		m["notes"] = *o.Notes
	}
	if o.Summary != nil {
		m["summary"] = *o.Summary
	}
	return m
}

func orgsToMaps(orgs []*model.Organization) []map[string]any {
	result := make([]map[string]any, len(orgs))
	for i, o := range orgs {
		result[i] = orgToMap(o)
	}
	return result
}

func registerOrgCommands(rootCmd *cobra.Command) {
	orgCmd := &cobra.Command{
		Use:   "org",
		Short: "Manage organizations",
	}

	orgCmd.AddCommand(orgAddCmd())
	orgCmd.AddCommand(orgListCmd())
	orgCmd.AddCommand(orgShowCmd())
	orgCmd.AddCommand(orgEditCmd())
	orgCmd.AddCommand(orgDeleteCmd())

	rootCmd.AddCommand(orgCmd)
}

func orgAddCmd() *cobra.Command {
	var domain, industry, notes string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			input := model.CreateOrgInput{
				Name:     args[0],
				Domain:   nilIfEmpty(domain),
				Industry: nilIfEmpty(industry),
				Notes:    nilIfEmpty(notes),
			}

			r := repo.NewOrgRepo(db)
			org, err := r.Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			data := []map[string]any{orgToMap(org)}
			return format.Output(os.Stdout, resolveFormat(), data, orgColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "domain name")
	cmd.Flags().StringVar(&industry, "industry", "", "industry")
	cmd.Flags().StringVar(&notes, "notes", "", "notes")

	return cmd
}

func orgListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewOrgRepo(db)
			orgs, err := r.FindAll(cmd.Context(), limit)
			if err != nil {
				return err
			}

			return format.Output(os.Stdout, resolveFormat(), orgsToMaps(orgs), orgColumns, flagQuiet)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "max results")

	return cmd
}

func orgShowCmd() *cobra.Command {
	var withPeople bool

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show an organization's details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid organization ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			orgRepo := repo.NewOrgRepo(db)
			org, err := orgRepo.FindByID(cmd.Context(), id)
			if err != nil {
				return err
			}

			data := orgToMap(org)

			if withPeople {
				people, err := orgRepo.FindPeople(cmd.Context(), id)
				if err != nil {
					return err
				}
				data["people"] = peopleToMaps(people)
			}

			return format.Output(os.Stdout, resolveFormat(), []map[string]any{data}, orgColumns, flagQuiet)
		},
	}

	cmd.Flags().BoolVar(&withPeople, "with-people", false, "include people in this organization")

	return cmd
}

func orgEditCmd() *cobra.Command {
	var name, domain, industry, notes string

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid organization ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			input := model.UpdateOrgInput{}
			if cmd.Flags().Changed("name") {
				input.Name = &name
			}
			if cmd.Flags().Changed("domain") {
				input.Domain = &domain
			}
			if cmd.Flags().Changed("industry") {
				input.Industry = &industry
			}
			if cmd.Flags().Changed("notes") {
				input.Notes = &notes
			}

			r := repo.NewOrgRepo(db)
			org, err := r.Update(cmd.Context(), id, input)
			if err != nil {
				return err
			}

			data := []map[string]any{orgToMap(org)}
			return format.Output(os.Stdout, resolveFormat(), data, orgColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "organization name")
	cmd.Flags().StringVar(&domain, "domain", "", "domain name")
	cmd.Flags().StringVar(&industry, "industry", "", "industry")
	cmd.Flags().StringVar(&notes, "notes", "", "notes")

	return cmd
}

func orgDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an organization (soft-delete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid organization ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewOrgRepo(db)
			if err := r.Archive(cmd.Context(), id); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Deleted organization #%d\n", id)
			return nil
		},
	}
}
