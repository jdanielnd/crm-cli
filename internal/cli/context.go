package cli

import (
	"fmt"
	"os"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/spf13/cobra"
)

func registerContextCommand(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "context <person_id_or_name>",
		Short: "Full context briefing for a person",
		Long:  "Returns person profile, organization, recent interactions, open deals, open tasks, relationships, and tags.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			pr := repo.NewPersonRepo(db)

			// Try to parse as ID first, otherwise search by name
			var person *model.Person
			id, parseErr := parseID(args[0])
			if parseErr == nil {
				person, err = pr.FindByID(cmd.Context(), id)
				if err != nil {
					return err
				}
			} else {
				people, err := pr.Search(cmd.Context(), args[0], 1)
				if err != nil {
					return err
				}
				if len(people) == 0 {
					return model.NewExitError(model.ErrNotFound, "no person found matching %q", args[0])
				}
				person = people[0]
			}

			ctx := cmd.Context()
			result := map[string]any{}

			// Person profile
			result["person"] = personToMap(person)

			// Organization
			if person.OrgID != nil {
				or := repo.NewOrgRepo(db)
				org, err := or.FindByID(ctx, *person.OrgID)
				if err == nil {
					result["organization"] = orgToMap(org)
				}
			}

			// Recent interactions
			ir := repo.NewInteractionRepo(db)
			interactions, err := ir.FindAll(ctx, model.InteractionFilters{PersonID: &person.ID, Limit: 10})
			if err == nil && len(interactions) > 0 {
				result["recent_interactions"] = interactionsToMaps(interactions)
			}

			// Open deals
			dr := repo.NewDealRepo(db)
			deals, err := dr.FindAll(ctx, model.DealFilters{PersonID: &person.ID})
			if err == nil && len(deals) > 0 {
				result["deals"] = dealsToMaps(deals)
			}

			// Open tasks
			tr := repo.NewTaskRepo(db)
			tasks, err := tr.FindAll(ctx, model.TaskFilters{PersonID: &person.ID})
			if err == nil && len(tasks) > 0 {
				result["tasks"] = tasksToMaps(tasks)
			}

			// Relationships
			rr := repo.NewRelationshipRepo(db)
			rels, err := rr.FindForPerson(ctx, person.ID)
			if err == nil && len(rels) > 0 {
				relMaps := make([]map[string]any, len(rels))
				for i, rel := range rels {
					relMaps[i] = map[string]any{
						"id":                rel.ID,
						"person_id":         rel.PersonID,
						"related_person_id": rel.RelatedPersonID,
						"type":              rel.Type,
					}
					if rel.Notes != nil {
						relMaps[i]["notes"] = *rel.Notes
					}
				}
				result["relationships"] = relMaps
			}

			// Tags
			tagRepo := repo.NewTagRepo(db)
			tags, err := tagRepo.GetForEntity(ctx, "person", person.ID)
			if err == nil && len(tags) > 0 {
				tagNames := make([]string, len(tags))
				for i, t := range tags {
					tagNames[i] = t.Name
				}
				result["tags"] = tagNames
			}

			// Output as a single-item array for consistency
			data := []map[string]any{result}
			cols := []format.ColumnDef{
				{Header: "Person", Field: "person"},
			}
			return format.Output(os.Stdout, resolveFormat(), data, cols, flagQuiet)
		},
	}

	rootCmd.AddCommand(cmd)
}

func parseID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
