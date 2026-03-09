package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/spf13/cobra"
)

var relationshipColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Person", Field: "person_id"},
	{Header: "Related", Field: "related_person_id"},
	{Header: "Type", Field: "type"},
	{Header: "Notes", Field: "notes"},
}

func registerRelateCommands(personCmd *cobra.Command) {
	personCmd.AddCommand(personRelateCmd())
	personCmd.AddCommand(personRelationshipsCmd())
	personCmd.AddCommand(personUnrelateCmd())
}

func personRelateCmd() *cobra.Command {
	var relType, notes string

	cmd := &cobra.Command{
		Use:   "relate <person_id> <related_person_id>",
		Short: "Create a relationship between two people",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			personID, err := parseEntityID(args[0], "person")
			if err != nil {
				return err
			}
			relatedID, err := parseEntityID(args[1], "person")
			if err != nil {
				return err
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewRelationshipRepo(db)
			rel, err := r.Create(cmd.Context(), personID, relatedID, relType, nilIfEmpty(notes))
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Related person #%d to person #%d as %q\n", rel.PersonID, rel.RelatedPersonID, rel.Type)
			return nil
		},
	}

	cmd.Flags().StringVar(&relType, "type", "colleague", "relationship type: "+strings.Join(model.RelationshipTypes, ", "))
	cmd.Flags().StringVar(&notes, "notes", "", "notes about the relationship")

	return cmd
}

func personRelationshipsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "relationships <person_id>",
		Aliases: []string{"rels"},
		Short:   "List relationships for a person",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			personID, err := parseEntityID(args[0], "person")
			if err != nil {
				return err
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewRelationshipRepo(db)
			rels, err := r.FindForPerson(cmd.Context(), personID)
			if err != nil {
				return err
			}

			data := make([]map[string]any, len(rels))
			for i, rel := range rels {
				m := map[string]any{
					"id":                rel.ID,
					"person_id":         rel.PersonID,
					"related_person_id": rel.RelatedPersonID,
					"type":              rel.Type,
					"created_at":        rel.CreatedAt,
				}
				if rel.Notes != nil {
					m["notes"] = *rel.Notes
				}
				data[i] = m
			}
			return format.Output(os.Stdout, resolveFormat(), data, relationshipColumns, flagQuiet)
		},
	}
}

func personUnrelateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unrelate <relationship_id>",
		Short: "Remove a relationship",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseEntityID(args[0], "relationship")
			if err != nil {
				return err
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewRelationshipRepo(db)
			if err := r.Delete(cmd.Context(), id); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Removed relationship #%d\n", id)
			return nil
		},
	}
}
