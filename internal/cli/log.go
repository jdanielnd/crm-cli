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

var interactionColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Type", Field: "type"},
	{Header: "Subject", Field: "subject"},
	{Header: "People", Field: "person_ids"},
	{Header: "Date", Field: "occurred_at"},
	{Header: "Direction", Field: "direction"},
}

func interactionToMap(i *model.Interaction) map[string]any {
	m := map[string]any{
		"id":          i.ID,
		"uuid":        i.UUID,
		"type":        i.Type,
		"occurred_at": i.OccurredAt,
		"person_ids":  i.PersonIDs,
		"created_at":  i.CreatedAt,
		"updated_at":  i.UpdatedAt,
	}
	if i.Subject != nil {
		m["subject"] = *i.Subject
	}
	if i.Content != nil {
		m["content"] = *i.Content
	}
	if i.Direction != nil {
		m["direction"] = *i.Direction
	}
	return m
}

func interactionsToMaps(interactions []*model.Interaction) []map[string]any {
	result := make([]map[string]any, len(interactions))
	for i, inter := range interactions {
		result[i] = interactionToMap(inter)
	}
	return result
}

func registerLogCommands(rootCmd *cobra.Command) {
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Log interactions (calls, emails, meetings, notes, messages)",
	}

	for _, iType := range model.InteractionTypes {
		logCmd.AddCommand(logTypeCmd(iType))
	}

	logCmd.AddCommand(logListCmd())

	rootCmd.AddCommand(logCmd)
}

func logTypeCmd(interactionType string) *cobra.Command {
	var subject, content, direction, at string

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <person_id> [person_id...]", interactionType),
		Short: fmt.Sprintf("Log a %s", interactionType),
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var personIDs []int64
			for _, arg := range args {
				id, err := strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return model.NewExitError(model.ErrValidation, "invalid person ID: %s", arg)
				}
				personIDs = append(personIDs, id)
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			if direction != "" && !model.ValidInteractionDirection(direction) {
				return model.NewExitError(model.ErrValidation, "invalid direction: %s (must be one of: %s)", direction, strings.Join(model.InteractionDirections, ", "))
			}

			input := model.CreateInteractionInput{
				Type:      interactionType,
				Subject:   nilIfEmpty(subject),
				Content:   nilIfEmpty(content),
				Direction: nilIfEmpty(direction),
				PersonIDs: personIDs,
			}
			if at != "" {
				input.OccurredAt = &at
			}

			r := repo.NewInteractionRepo(db)
			interaction, err := r.Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			data := []map[string]any{interactionToMap(interaction)}
			return format.Output(os.Stdout, resolveFormat(), data, interactionColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "subject/title")
	cmd.Flags().StringVar(&content, "content", "", "content/body")
	cmd.Flags().StringVar(&direction, "direction", "", "direction: inbound or outbound")
	cmd.Flags().StringVar(&at, "at", "", "when the interaction occurred (ISO 8601)")

	return cmd
}

func logListCmd() *cobra.Command {
	var personID int64
	var iType string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List interactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			filters := model.InteractionFilters{Limit: limit}
			if cmd.Flags().Changed("person") {
				filters.PersonID = &personID
			}
			if iType != "" {
				filters.Type = &iType
			}

			r := repo.NewInteractionRepo(db)
			interactions, err := r.FindAll(cmd.Context(), filters)
			if err != nil {
				return err
			}

			return format.Output(os.Stdout, resolveFormat(), interactionsToMaps(interactions), interactionColumns, flagQuiet)
		},
	}

	cmd.Flags().Int64Var(&personID, "person", 0, "filter by person ID")
	cmd.Flags().StringVar(&iType, "type", "", "filter by interaction type")
	cmd.Flags().IntVar(&limit, "limit", 0, "max results")

	return cmd
}
