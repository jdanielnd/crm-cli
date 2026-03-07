package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jdanielnd/crm-cli/internal/model"
)

// OrgRepo handles organization database operations.
type OrgRepo struct {
	db *sql.DB
}

// NewOrgRepo creates a new OrgRepo.
func NewOrgRepo(db *sql.DB) *OrgRepo {
	return &OrgRepo{db: db}
}

var orgColumns = "id, uuid, name, domain, industry, notes, summary, created_at, updated_at"

func scanOrg(row interface{ Scan(...any) error }) (*model.Organization, error) {
	var o model.Organization
	err := row.Scan(
		&o.ID, &o.UUID, &o.Name, &o.Domain, &o.Industry,
		&o.Notes, &o.Summary, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// Create inserts a new organization.
func (r *OrgRepo) Create(ctx context.Context, input model.CreateOrgInput) (*model.Organization, error) {
	id := uuid.New().String()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO organizations (uuid, name, domain, industry, notes)
		 VALUES (?, ?, ?, ?, ?)`,
		id, input.Name, input.Domain, input.Industry, input.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("insert organization: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	return r.FindByID(ctx, rowID)
}

// FindByID returns an organization by ID.
func (r *OrgRepo) FindByID(ctx context.Context, id int64) (*model.Organization, error) {
	row := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT %s FROM organizations WHERE id = ? AND archived = 0", orgColumns), id)
	o, err := scanOrg(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization %d: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find organization %d: %w", id, err)
	}
	return o, nil
}

// FindAll lists organizations with optional filters.
func (r *OrgRepo) FindAll(ctx context.Context, limit int) ([]*model.Organization, error) {
	query := fmt.Sprintf("SELECT %s FROM organizations WHERE archived = 0 ORDER BY updated_at DESC", orgColumns)
	var args []any
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		orgs = append(orgs, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organizations: %w", err)
	}
	return orgs, nil
}

// Update modifies an organization's fields.
func (r *OrgRepo) Update(ctx context.Context, id int64, input model.UpdateOrgInput) (*model.Organization, error) {
	var setClauses []string
	var args []any

	if input.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *input.Name)
	}
	if input.Domain != nil {
		setClauses = append(setClauses, "domain = ?")
		args = append(args, *input.Domain)
	}
	if input.Industry != nil {
		setClauses = append(setClauses, "industry = ?")
		args = append(args, *input.Industry)
	}
	if input.Notes != nil {
		setClauses = append(setClauses, "notes = ?")
		args = append(args, *input.Notes)
	}
	if input.Summary != nil {
		setClauses = append(setClauses, "summary = ?")
		args = append(args, *input.Summary)
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = datetime('now')")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE organizations SET %s WHERE id = ? AND archived = 0", strings.Join(setClauses, ", "))
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update organization %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("organization %d: %w", id, model.ErrNotFound)
	}

	return r.FindByID(ctx, id)
}

// Archive soft-deletes an organization.
func (r *OrgRepo) Archive(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE organizations SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0", id)
	if err != nil {
		return fmt.Errorf("archive organization %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("organization %d: %w", id, model.ErrNotFound)
	}

	return nil
}

// Search performs full-text search on organizations.
func (r *OrgRepo) Search(ctx context.Context, query string, limit int) ([]*model.Organization, error) {
	if limit <= 0 {
		limit = 20
	}

	ftsQuery := `"` + strings.ReplaceAll(query, `"`, `""`) + `"` + "*"
	sqlStr := fmt.Sprintf(
		`SELECT %s FROM organizations WHERE archived = 0 AND id IN (SELECT rowid FROM organizations_fts WHERE organizations_fts MATCH ?) ORDER BY updated_at DESC LIMIT ?`,
		orgColumns,
	)

	rows, err := r.db.QueryContext(ctx, sqlStr, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		orgs = append(orgs, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}
	return orgs, nil
}

// FindPeople returns all people belonging to an organization.
func (r *OrgRepo) FindPeople(ctx context.Context, orgID int64) ([]*model.Person, error) {
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM people WHERE org_id = ? AND archived = 0 ORDER BY first_name", personColumns), orgID)
	if err != nil {
		return nil, fmt.Errorf("find people for org %d: %w", orgID, err)
	}
	defer rows.Close()

	var people []*model.Person
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, fmt.Errorf("scan person: %w", err)
		}
		people = append(people, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate people: %w", err)
	}
	return people, nil
}
