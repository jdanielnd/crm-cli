package cli

import (
	"fmt"
	"os"
	"text/template"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/spf13/cobra"
)

var templateColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Name", Field: "name"},
	{Header: "Created", Field: "created_at"},
}

func templateToMap(t *model.Template) map[string]any {
	return map[string]any{
		"id":         t.ID,
		"uuid":       t.UUID,
		"name":       t.Name,
		"body":       t.Body,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
}

func templatesToMaps(templates []*model.Template) []map[string]any {
	result := make([]map[string]any, len(templates))
	for i, t := range templates {
		result[i] = templateToMap(t)
	}
	return result
}

func registerTemplateCommands(rootCmd *cobra.Command) {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Manage message templates",
	}

	templateCmd.AddCommand(templateAddCmd())
	templateCmd.AddCommand(templateListCmd())
	templateCmd.AddCommand(templateShowCmd())
	templateCmd.AddCommand(templateDeleteCmd())
	templateCmd.AddCommand(templateRenderCmd())

	rootCmd.AddCommand(templateCmd)
}

func templateAddCmd() *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return model.NewExitError(model.ErrValidation, "--body is required")
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTemplateRepo(db)
			tmpl, err := r.Create(cmd.Context(), model.CreateTemplateInput{
				Name: args[0],
				Body: body,
			})
			if err != nil {
				return err
			}

			data := []map[string]any{templateToMap(tmpl)}
			return format.Output(os.Stdout, resolveFormat(), data, templateColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "template body (Go text/template syntax)")
	return cmd
}

func templateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTemplateRepo(db)
			templates, err := r.FindAll(cmd.Context())
			if err != nil {
				return err
			}

			return format.Output(os.Stdout, resolveFormat(), templatesToMaps(templates), templateColumns, flagQuiet)
		},
	}
}

func templateShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show a template's body",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTemplateRepo(db)
			tmpl, err := r.FindByName(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(os.Stdout, tmpl.Body)
			return nil
		},
	}
}

func templateDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTemplateRepo(db)
			if err := r.Delete(cmd.Context(), args[0]); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Deleted template %q\n", args[0])
			return nil
		},
	}
}

func templateRenderCmd() *cobra.Command {
	var personID int64

	cmd := &cobra.Command{
		Use:   "render <name>",
		Short: "Render a template with person data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("person") {
				return model.NewExitError(model.ErrValidation, "--person is required")
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			// Fetch template
			tr := repo.NewTemplateRepo(db)
			tmpl, err := tr.FindByName(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			// Fetch person
			pr := repo.NewPersonRepo(db)
			person, err := pr.FindByID(cmd.Context(), personID)
			if err != nil {
				return err
			}

			// Build template data
			data := model.TemplateData{
				FirstName: person.FirstName,
			}
			if person.LastName != nil {
				data.LastName = *person.LastName
			}
			if person.Email != nil {
				data.Email = *person.Email
			}
			if person.Phone != nil {
				data.Phone = *person.Phone
			}
			if person.Title != nil {
				data.Title = *person.Title
			}
			if person.Company != nil {
				data.Company = *person.Company
			}
			if person.Location != nil {
				data.Location = *person.Location
			}

			// Fetch org if linked
			if person.OrgID != nil {
				or := repo.NewOrgRepo(db)
				org, err := or.FindByID(cmd.Context(), *person.OrgID)
				if err == nil {
					data.OrgName = org.Name
					if org.Domain != nil {
						data.OrgDomain = *org.Domain
					}
					if org.Industry != nil {
						data.OrgIndustry = *org.Industry
					}
				}
			}

			// Parse and execute template
			t, err := template.New(tmpl.Name).Parse(tmpl.Body)
			if err != nil {
				return model.NewExitError(model.ErrValidation, "invalid template syntax: %v", err)
			}

			return t.Execute(os.Stdout, data)
		},
	}

	cmd.Flags().Int64Var(&personID, "person", 0, "person ID for template data")
	return cmd
}
