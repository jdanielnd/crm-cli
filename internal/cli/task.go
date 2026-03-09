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

var taskColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Title", Field: "title"},
	{Header: "Due", Field: "due_at"},
	{Header: "Priority", Field: "priority"},
	{Header: "Done", Field: "completed"},
	{Header: "Person", Field: "person_id"},
}

func taskToMap(t *model.Task) map[string]any {
	m := map[string]any{
		"id":         t.ID,
		"uuid":       t.UUID,
		"title":      t.Title,
		"priority":   t.Priority,
		"completed":  t.Completed,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
	if t.Description != nil {
		m["description"] = *t.Description
	}
	if t.PersonID != nil {
		m["person_id"] = *t.PersonID
	}
	if t.DealID != nil {
		m["deal_id"] = *t.DealID
	}
	if t.DueAt != nil {
		m["due_at"] = *t.DueAt
	}
	if t.CompletedAt != nil {
		m["completed_at"] = *t.CompletedAt
	}
	return m
}

func tasksToMaps(tasks []*model.Task) []map[string]any {
	result := make([]map[string]any, len(tasks))
	for i, t := range tasks {
		result[i] = taskToMap(t)
	}
	return result
}

func registerTaskCommands(rootCmd *cobra.Command) {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks and follow-ups",
	}

	taskCmd.AddCommand(taskAddCmd())
	taskCmd.AddCommand(taskListCmd())
	taskCmd.AddCommand(taskShowCmd())
	taskCmd.AddCommand(taskEditCmd())
	taskCmd.AddCommand(taskDoneCmd())
	taskCmd.AddCommand(taskDeleteCmd())

	rootCmd.AddCommand(taskCmd)
}

func taskAddCmd() *cobra.Command {
	var description, due, priority string
	var personID, dealID int64

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			input := model.CreateTaskInput{
				Title:       args[0],
				Description: nilIfEmpty(description),
				Priority:    priority,
				DueAt:       nilIfEmpty(due),
			}
			if cmd.Flags().Changed("person") {
				input.PersonID = &personID
			}
			if cmd.Flags().Changed("deal") {
				input.DealID = &dealID
			}

			r := repo.NewTaskRepo(db)
			task, err := r.Create(cmd.Context(), input)
			if err != nil {
				return err
			}

			data := []map[string]any{taskToMap(task)}
			return format.Output(os.Stdout, resolveFormat(), data, taskColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "task description")
	cmd.Flags().StringVar(&due, "due", "", "due date (ISO 8601)")
	cmd.Flags().StringVar(&priority, "priority", "medium", "priority: "+strings.Join(model.Priorities, ", "))
	cmd.Flags().Int64Var(&personID, "person", 0, "associated person ID")
	cmd.Flags().Int64Var(&dealID, "deal", 0, "associated deal ID")

	return cmd
}

func taskListCmd() *cobra.Command {
	var personID, dealID int64
	var overdue, all bool
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			filters := model.TaskFilters{
				Limit:            limit,
				Overdue:          overdue,
				IncludeCompleted: all,
			}
			if cmd.Flags().Changed("person") {
				filters.PersonID = &personID
			}
			if cmd.Flags().Changed("deal") {
				filters.DealID = &dealID
			}

			r := repo.NewTaskRepo(db)
			tasks, err := r.FindAll(cmd.Context(), filters)
			if err != nil {
				return err
			}

			return format.Output(os.Stdout, resolveFormat(), tasksToMaps(tasks), taskColumns, flagQuiet)
		},
	}

	cmd.Flags().Int64Var(&personID, "person", 0, "filter by person ID")
	cmd.Flags().Int64Var(&dealID, "deal", 0, "filter by deal ID")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "show only overdue tasks")
	cmd.Flags().BoolVar(&all, "all", false, "include completed tasks")
	cmd.Flags().IntVar(&limit, "limit", 0, "max results")

	return cmd
}

func taskShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseEntityID(args[0], "task")
			if err != nil {
				return err
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTaskRepo(db)
			task, err := r.FindByID(cmd.Context(), id)
			if err != nil {
				return err
			}

			data := []map[string]any{taskToMap(task)}
			return format.Output(os.Stdout, resolveFormat(), data, taskColumns, flagQuiet)
		},
	}
}

func taskEditCmd() *cobra.Command {
	var title, description, due, priority string
	var personID, dealID int64

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseEntityID(args[0], "task")
			if err != nil {
				return err
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			input := model.UpdateTaskInput{}
			if cmd.Flags().Changed("title") {
				input.Title = &title
			}
			if cmd.Flags().Changed("description") {
				input.Description = &description
			}
			if cmd.Flags().Changed("due") {
				input.DueAt = &due
			}
			if cmd.Flags().Changed("priority") {
				input.Priority = &priority
			}
			if cmd.Flags().Changed("person") {
				input.PersonID = &personID
			}
			if cmd.Flags().Changed("deal") {
				input.DealID = &dealID
			}

			r := repo.NewTaskRepo(db)
			task, err := r.Update(cmd.Context(), id, input)
			if err != nil {
				return err
			}

			data := []map[string]any{taskToMap(task)}
			return format.Output(os.Stdout, resolveFormat(), data, taskColumns, flagQuiet)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "task title")
	cmd.Flags().StringVar(&description, "description", "", "task description")
	cmd.Flags().StringVar(&due, "due", "", "due date (ISO 8601)")
	cmd.Flags().StringVar(&priority, "priority", "", "priority: "+strings.Join(model.Priorities, ", "))
	cmd.Flags().Int64Var(&personID, "person", 0, "associated person ID")
	cmd.Flags().Int64Var(&dealID, "deal", 0, "associated deal ID")

	return cmd
}

func taskDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "done <id>",
		Short: "Mark a task as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseEntityID(args[0], "task")
			if err != nil {
				return err
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTaskRepo(db)
			task, err := r.Complete(cmd.Context(), id)
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Completed task #%d: %s\n", task.ID, task.Title)
			return nil
		},
	}
}

func taskDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task (soft-delete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseEntityID(args[0], "task")
			if err != nil {
				return err
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			r := repo.NewTaskRepo(db)
			if err := r.Archive(cmd.Context(), id); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Deleted task #%d\n", id)
			return nil
		},
	}
}
