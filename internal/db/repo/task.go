package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jdanielnd/crm-cli/internal/model"
)

// TaskRepo handles task database operations.
type TaskRepo struct {
	db *sql.DB
}

// NewTaskRepo creates a new TaskRepo.
func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func scanTask(row interface{ Scan(...any) error }) (*model.Task, error) {
	var t model.Task
	err := row.Scan(
		&t.ID, &t.UUID, &t.Title, &t.Description, &t.PersonID, &t.DealID,
		&t.DueAt, &t.Priority, &t.Completed, &t.CompletedAt,
		&t.Archived, &t.CreatedAt, &t.UpdatedAt,
	)
	return &t, err
}

// Create inserts a new task.
func (r *TaskRepo) Create(ctx context.Context, input model.CreateTaskInput) (*model.Task, error) {
	if !model.ValidPriority(input.Priority) {
		return nil, fmt.Errorf("invalid priority %q: %w", input.Priority, model.ErrValidation)
	}

	id := uuid.New().String()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (uuid, title, description, person_id, deal_id, due_at, priority)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, input.Title, input.Description, input.PersonID, input.DealID, input.DueAt, input.Priority)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	taskID, _ := result.LastInsertId()
	return r.FindByID(ctx, taskID)
}

// FindByID returns a task by ID.
func (r *TaskRepo) FindByID(ctx context.Context, id int64) (*model.Task, error) {
	t, err := scanTask(r.db.QueryRowContext(ctx,
		`SELECT id, uuid, title, description, person_id, deal_id, due_at, priority,
		        completed, completed_at, archived, created_at, updated_at
		 FROM tasks WHERE id = ? AND archived = 0`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task %d: %w", id, model.ErrNotFound)
		}
		return nil, fmt.Errorf("find task %d: %w", id, err)
	}
	return t, nil
}

// FindAll returns tasks with optional filters.
func (r *TaskRepo) FindAll(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error) {
	query := `SELECT id, uuid, title, description, person_id, deal_id, due_at, priority,
	                 completed, completed_at, archived, created_at, updated_at
	          FROM tasks WHERE archived = 0`
	var args []any

	if filters.PersonID != nil {
		query += " AND person_id = ?"
		args = append(args, *filters.PersonID)
	}
	if filters.DealID != nil {
		query += " AND deal_id = ?"
		args = append(args, *filters.DealID)
	}
	if filters.Overdue {
		query += " AND completed = 0 AND due_at IS NOT NULL AND due_at < datetime('now')"
	}
	if !filters.IncludeCompleted {
		query += " AND completed = 0"
	}

	query += " ORDER BY CASE WHEN due_at IS NULL THEN 1 ELSE 0 END, due_at ASC, created_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// Complete marks a task as completed.
func (r *TaskRepo) Complete(ctx context.Context, id int64) (*model.Task, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET completed = 1, completed_at = datetime('now'), updated_at = datetime('now')
		 WHERE id = ? AND archived = 0 AND completed = 0`, id)
	if err != nil {
		return nil, fmt.Errorf("complete task %d: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		// Check if the task exists but is already completed
		var completed int
		err := r.db.QueryRowContext(ctx, "SELECT completed FROM tasks WHERE id = ? AND archived = 0", id).Scan(&completed)
		if err != nil {
			return nil, fmt.Errorf("task %d: %w", id, model.ErrNotFound)
		}
		if completed == 1 {
			return nil, fmt.Errorf("task %d is already completed: %w", id, model.ErrConflict)
		}
		return nil, fmt.Errorf("task %d: %w", id, model.ErrNotFound)
	}
	return r.FindByID(ctx, id)
}

// Update modifies a task.
func (r *TaskRepo) Update(ctx context.Context, id int64, input model.UpdateTaskInput) (*model.Task, error) {
	var setClauses []string
	var args []any

	if input.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *input.Title)
	}
	if input.Description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *input.Description)
	}
	if input.PersonID != nil {
		setClauses = append(setClauses, "person_id = ?")
		args = append(args, *input.PersonID)
	}
	if input.DealID != nil {
		setClauses = append(setClauses, "deal_id = ?")
		args = append(args, *input.DealID)
	}
	if input.DueAt != nil {
		setClauses = append(setClauses, "due_at = ?")
		args = append(args, *input.DueAt)
	}
	if input.Priority != nil {
		if !model.ValidPriority(*input.Priority) {
			return nil, fmt.Errorf("invalid priority %q: %w", *input.Priority, model.ErrValidation)
		}
		setClauses = append(setClauses, "priority = ?")
		args = append(args, *input.Priority)
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = datetime('now')")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ? AND archived = 0", strings.Join(setClauses, ", "))
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update task %d: %w", id, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("task %d: %w", id, model.ErrNotFound)
	}

	return r.FindByID(ctx, id)
}

// Archive soft-deletes a task.
func (r *TaskRepo) Archive(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0", id)
	if err != nil {
		return fmt.Errorf("archive task %d: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %d: %w", id, model.ErrNotFound)
	}
	return nil
}
