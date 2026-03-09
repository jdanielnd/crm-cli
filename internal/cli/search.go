package cli

import (
	"os"
	"strings"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/spf13/cobra"
)

var searchColumns = []format.ColumnDef{
	{Header: "Type", Field: "type"},
	{Header: "ID", Field: "id"},
	{Header: "Name", Field: "name"},
	{Header: "Detail", Field: "detail"},
}

func registerSearchCommand(rootCmd *cobra.Command) {
	var entityType string
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search across all entities",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			query := args[0]

			if entityType != "" && !model.ValidEntityType(entityType) {
				return model.NewExitError(model.ErrValidation, "invalid entity type: %s (must be one of: %s)", entityType, strings.Join(model.EntityTypes, ", "))
			}

			var data []map[string]any

			if entityType == "" || entityType == "person" {
				pr := repo.NewPersonRepo(db)
				people, err := pr.Search(cmd.Context(), query, limit)
				if err != nil {
					return err
				}
				for _, p := range people {
					name := p.FirstName
					if p.LastName != nil {
						name += " " + *p.LastName
					}
					detail := ""
					if p.Email != nil {
						detail = *p.Email
					}
					data = append(data, map[string]any{
						"type":   "person",
						"id":     p.ID,
						"name":   name,
						"detail": detail,
					})
				}
			}

			if entityType == "" || entityType == "organization" {
				or := repo.NewOrgRepo(db)
				orgs, err := or.Search(cmd.Context(), query, limit)
				if err != nil {
					return err
				}
				for _, o := range orgs {
					detail := ""
					if o.Domain != nil {
						detail = *o.Domain
					}
					data = append(data, map[string]any{
						"type":   "organization",
						"id":     o.ID,
						"name":   o.Name,
						"detail": detail,
					})
				}
			}

			if entityType == "" || entityType == "interaction" {
				ir := repo.NewInteractionRepo(db)
				interactions, err := ir.Search(cmd.Context(), query, limit)
				if err != nil {
					return err
				}
				for _, i := range interactions {
					name := i.Type
					if i.Subject != nil {
						name += ": " + *i.Subject
					}
					data = append(data, map[string]any{
						"type":   "interaction",
						"id":     i.ID,
						"name":   name,
						"detail": i.OccurredAt,
					})
				}
			}

			if entityType == "" || entityType == "deal" {
				dr := repo.NewDealRepo(db)
				deals, err := dr.Search(cmd.Context(), query, limit)
				if err != nil {
					return err
				}
				for _, d := range deals {
					detail := d.Stage
					data = append(data, map[string]any{
						"type":   "deal",
						"id":     d.ID,
						"name":   d.Title,
						"detail": detail,
					})
				}
			}

			return format.Output(os.Stdout, resolveFormat(), data, searchColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&entityType, "type", "", "filter by entity type: person, organization, interaction, deal")
	cmd.Flags().IntVar(&limit, "limit", 20, "max results per entity type")

	rootCmd.AddCommand(cmd)
}
