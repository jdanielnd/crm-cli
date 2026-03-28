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

// TemplateRepo handles template database operations.
type TemplateRepo struct {
	db *sql.DB
}

// NewTemplateRepo creates a new TemplateRepo.
func NewTemplateRepo(db *sql.DB) *TemplateRepo {
	return &TemplateRepo{db: db}
}

func scanTemplate(row interface{ Scan(...any) error }) (*model.Template, error) {
	var t model.Template
	err := row.Scan(&t.ID, &t.UUID, &t.Name, &t.Body, &t.CreatedAt, &t.UpdatedAt)
	return &t, err
}

// Create inserts a new template.
func (r *TemplateRepo) Create(ctx context.Context, input model.CreateTemplateInput) (*model.Template, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("template name is required: %w", model.ErrValidation)
	}
	if input.Body == "" {
		return nil, fmt.Errorf("template body is required: %w", model.ErrValidation)
	}

	id := uuid.New().String()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO templates (uuid, name, body) VALUES (?, ?, ?)`,
		id, input.Name, input.Body)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("template %q already exists: %w", input.Name, model.ErrConflict)
		}
		return nil, fmt.Errorf("create template: %w", err)
	}

	rowID, _ := result.LastInsertId()
	return r.FindByID(ctx, rowID)
}

// FindByID returns a template by ID.
func (r *TemplateRepo) FindByID(ctx context.Context, id int64) (*model.Template, error) {
	t, err := scanTemplate(r.db.QueryRowContext(ctx,
		`SELECT id, uuid, name, body, created_at, updated_at FROM templates WHERE id = ?`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("template %d: %w", id, model.ErrNotFound)
		}
		return nil, fmt.Errorf("find template %d: %w", id, err)
	}
	return t, nil
}

// FindByName returns a template by name.
func (r *TemplateRepo) FindByName(ctx context.Context, name string) (*model.Template, error) {
	t, err := scanTemplate(r.db.QueryRowContext(ctx,
		`SELECT id, uuid, name, body, created_at, updated_at FROM templates WHERE name = ?`, name))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("template %q: %w", name, model.ErrNotFound)
		}
		return nil, fmt.Errorf("find template %q: %w", name, err)
	}
	return t, nil
}

// FindAll returns all templates.
func (r *TemplateRepo) FindAll(ctx context.Context) ([]*model.Template, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, uuid, name, body, created_at, updated_at FROM templates ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var templates []*model.Template
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// Delete removes a template by name.
func (r *TemplateRepo) Delete(ctx context.Context, name string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM templates WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("template %q: %w", name, model.ErrNotFound)
	}
	return nil
}

// isUniqueViolation checks if an error is a SQLite UNIQUE constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "unique constraint")
}
