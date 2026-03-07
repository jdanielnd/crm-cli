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

var personColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Name", Field: "name"},
	{Header: "Email", Field: "email"},
	{Header: "Phone", Field: "phone"},
	{Header: "Company", Field: "company"},
	{Header: "Location", Field: "location"},
}

func personToMap(p *model.Person) map[string]any {
	name := p.FirstName
	if p.LastName != nil {
		name += " " + *p.LastName
	}
	m := map[string]any{
		"id":         p.ID,
		"uuid":       p.UUID,
		"first_name": p.FirstName,
		"name":       name,
		"created_at": p.CreatedAt,
		"updated_at": p.UpdatedAt,
	}
	if p.LastName != nil {
		m["last_name"] = *p.LastName
	}
	if p.Email != nil {
		m["email"] = *p.Email
	}
	if p.Phone != nil {
		m["phone"] = *p.Phone
	}
	if p.Title != nil {
		m["title"] = *p.Title
	}
	if p.Company != nil {
		m["company"] = *p.Company
	}
	if p.Location != nil {
		m["location"] = *p.Location
	}
	if p.Notes != nil {
		m["notes"] = *p.Notes
	}
	if p.Summary != nil {
		m["summary"] = *p.Summary
	}
	if p.OrgID != nil {
		m["org_id"] = *p.OrgID
	}
	return m
}

func peopleToMaps(people []*model.Person) []map[string]any {
	result := make([]map[string]any, len(people))
	for i, p := range people {
		result[i] = personToMap(p)
	}
	return result
}

func registerPersonCommands(rootCmd *cobra.Command) {
	personCmd := &cobra.Command{
		Use:   "person",
		Short: "Manage contacts",
	}

	personCmd.AddCommand(personAddCmd())
	personCmd.AddCommand(personListCmd())
	personCmd.AddCommand(personShowCmd())
	personCmd.AddCommand(personEditCmd())
	personCmd.AddCommand(personDeleteCmd())
	registerRelateCommands(personCmd)

	rootCmd.AddCommand(personCmd)
}

func personAddCmd() *cobra.Command {
	var email, phone, title, company, location, notes string
	var orgID int64

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new person",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			firstName, lastName := parseName(args[0])
			input := model.CreatePersonInput{
				FirstName: firstName,
				LastName:  nilIfEmpty(lastName),
				Email:     nilIfEmpty(email),
				Phone:     nilIfEmpty(phone),
				Title:     nilIfEmpty(title),
				Company:   nilIfEmpty(company),
				Location:  nilIfEmpty(location),
				Notes:     nilIfEmpty(notes),
			}
			if cmd.Flags().Changed("org") {
				input.OrgID = &orgID
			}

			r := repo.NewPersonRepo(db)
			person, err := r.Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			data := []map[string]any{personToMap(person)}
			return format.Output(os.Stdout, resolveFormat(), data, personColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "email address")
	cmd.Flags().StringVar(&phone, "phone", "", "phone number")
	cmd.Flags().StringVar(&title, "title", "", "job title")
	cmd.Flags().StringVar(&company, "company", "", "company name")
	cmd.Flags().StringVar(&location, "location", "", "location")
	cmd.Flags().StringVar(&notes, "notes", "", "notes")
	cmd.Flags().Int64Var(&orgID, "org", 0, "organization ID")

	return cmd
}

func personListCmd() *cobra.Command {
	var tag string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all people",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			filters := model.PersonFilters{Limit: limit}
			if tag != "" {
				filters.Tag = &tag
			}

			r := repo.NewPersonRepo(db)
			people, err := r.FindAll(cmd.Context(), filters)
			if err != nil {
				return err
			}

			return format.Output(os.Stdout, resolveFormat(), peopleToMaps(people), personColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "filter by tag")
	cmd.Flags().IntVar(&limit, "limit", 0, "max results")

	return cmd
}

func personShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a person's details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid person ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewPersonRepo(db)
			person, err := r.FindByID(cmd.Context(), id)
			if err != nil {
				return err
			}

			data := []map[string]any{personToMap(person)}
			return format.Output(os.Stdout, resolveFormat(), data, personColumns, flagQuiet)
		},
	}
}

func personEditCmd() *cobra.Command {
	var firstName, lastName, email, phone, title, company, location, notes string
	var orgID int64

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a person",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid person ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			input := model.UpdatePersonInput{}
			if cmd.Flags().Changed("first-name") {
				input.FirstName = &firstName
			}
			if cmd.Flags().Changed("last-name") {
				input.LastName = &lastName
			}
			if cmd.Flags().Changed("email") {
				input.Email = &email
			}
			if cmd.Flags().Changed("phone") {
				input.Phone = &phone
			}
			if cmd.Flags().Changed("title") {
				input.Title = &title
			}
			if cmd.Flags().Changed("company") {
				input.Company = &company
			}
			if cmd.Flags().Changed("location") {
				input.Location = &location
			}
			if cmd.Flags().Changed("notes") {
				input.Notes = &notes
			}
			if cmd.Flags().Changed("org") {
				input.OrgID = &orgID
			}

			r := repo.NewPersonRepo(db)
			person, err := r.Update(cmd.Context(), id, input)
			if err != nil {
				return err
			}

			data := []map[string]any{personToMap(person)}
			return format.Output(os.Stdout, resolveFormat(), data, personColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&firstName, "first-name", "", "first name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "last name")
	cmd.Flags().StringVar(&email, "email", "", "email address")
	cmd.Flags().StringVar(&phone, "phone", "", "phone number")
	cmd.Flags().StringVar(&title, "title", "", "job title")
	cmd.Flags().StringVar(&company, "company", "", "company name")
	cmd.Flags().StringVar(&location, "location", "", "location")
	cmd.Flags().StringVar(&notes, "notes", "", "notes")
	cmd.Flags().Int64Var(&orgID, "org", 0, "organization ID")

	return cmd
}

func personDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a person (soft-delete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid person ID: %s", args[0])
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewPersonRepo(db)
			if err := r.Archive(cmd.Context(), id); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Deleted person #%d\n", id)
			return nil
		},
	}
}

// parseName splits "Jane Smith" into first and last.
func parseName(name string) (string, string) {
	parts := splitName(name)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], joinParts(parts[1:])
}

func splitName(name string) []string {
	var parts []string
	for _, p := range splitWhitespace(name) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitWhitespace(s string) []string {
	var result []string
	current := ""
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
