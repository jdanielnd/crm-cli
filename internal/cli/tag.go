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

var tagColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Name", Field: "name"},
}

func registerTagCommands(rootCmd *cobra.Command) {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}

	tagCmd.AddCommand(tagListCmd())
	tagCmd.AddCommand(tagApplyCmd())
	tagCmd.AddCommand(tagRemoveCmd())
	tagCmd.AddCommand(tagShowCmd())
	tagCmd.AddCommand(tagDeleteCmd())

	rootCmd.AddCommand(tagCmd)
}

func tagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTagRepo(db)
			tags, err := r.FindAll(cmd.Context())
			if err != nil {
				return err
			}

			data := make([]map[string]any, len(tags))
			for i, t := range tags {
				data[i] = map[string]any{"id": t.ID, "name": t.Name}
			}
			return format.Output(os.Stdout, resolveFormat(), data, tagColumns, flagQuiet)
		},
	}
}

func tagApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply <entity_type> <entity_id> <tag>",
		Short: "Apply a tag to an entity",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			entityType := args[0]
			entityID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid entity ID: %s", args[1])
			}
			tagName := args[2]

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTagRepo(db)
			if err := r.Apply(cmd.Context(), entityType, entityID, tagName); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Tagged %s #%d with %q\n", entityType, entityID, tagName)
			return nil
		},
	}
}

func tagRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <entity_type> <entity_id> <tag>",
		Short: "Remove a tag from an entity",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			entityType := args[0]
			entityID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid entity ID: %s", args[1])
			}
			tagName := args[2]

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTagRepo(db)
			if err := r.Remove(cmd.Context(), entityType, entityID, tagName); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Removed tag %q from %s #%d\n", tagName, entityType, entityID)
			return nil
		},
	}
}

func tagShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <entity_type> <entity_id>",
		Short: "Show tags for an entity",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			entityType := args[0]
			entityID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid entity ID: %s", args[1])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTagRepo(db)
			tags, err := r.GetForEntity(cmd.Context(), entityType, entityID)
			if err != nil {
				return err
			}

			data := make([]map[string]any, len(tags))
			for i, t := range tags {
				data[i] = map[string]any{"id": t.ID, "name": t.Name}
			}
			return format.Output(os.Stdout, resolveFormat(), data, tagColumns, flagQuiet)
		},
	}
}

func tagDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <tag>",
		Short: "Delete a tag and all its associations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTagRepo(db)
			if err := r.Delete(cmd.Context(), args[0]); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Deleted tag %q\n", args[0])
			return nil
		},
	}
}
