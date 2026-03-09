package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jdanielnd/crm-cli/internal/model"
)

// PersonRepo handles person database operations.
type PersonRepo struct {
	db *sql.DB
}

// NewPersonRepo creates a new PersonRepo.
func NewPersonRepo(db *sql.DB) *PersonRepo {
	return &PersonRepo{db: db}
}

var personColumns = "id, uuid, first_name, last_name, email, phone, title, company, location, notes, summary, org_id, created_at, updated_at"

func scanPerson(row interface{ Scan(...any) error }) (*model.Person, error) {
	var p model.Person
	err := row.Scan(
		&p.ID, &p.UUID, &p.FirstName, &p.LastName,
		&p.Email, &p.Phone, &p.Title, &p.Company,
		&p.Location, &p.Notes, &p.Summary, &p.OrgID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Create inserts a new person after checking for duplicates.
func (r *PersonRepo) Create(ctx context.Context, input model.CreatePersonInput) (*model.Person, error) {
	// Check for duplicate: same first_name + last_name, or same email
	if err := r.checkDuplicate(ctx, input); err != nil {
		return nil, err
	}

	id := uuid.New().String()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO people (uuid, first_name, last_name, email, phone, title, company, location, notes, org_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, input.FirstName, input.LastName, input.Email, input.Phone,
		input.Title, input.Company, input.Location, input.Notes, input.OrgID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert person: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	return r.FindByID(ctx, rowID)
}

func (r *PersonRepo) checkDuplicate(ctx context.Context, input model.CreatePersonInput) error {
	// Check by email if provided
	if input.Email != nil && *input.Email != "" {
		var count int
		err := r.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM people WHERE email = ? AND archived = 0", *input.Email).Scan(&count)
		if err != nil {
			return fmt.Errorf("check duplicate email: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("a person with email %q already exists: %w", *input.Email, model.ErrConflict)
		}
	}

	// Check by full name if last name is provided
	if input.LastName != nil && *input.LastName != "" {
		var count int
		err := r.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM people WHERE first_name = ? AND last_name = ? AND archived = 0",
			input.FirstName, *input.LastName).Scan(&count)
		if err != nil {
			return fmt.Errorf("check duplicate name: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("a person named %q %q already exists: %w", input.FirstName, *input.LastName, model.ErrConflict)
		}
	}

	return nil
}

// FindByID returns a person by ID.
func (r *PersonRepo) FindByID(ctx context.Context, id int64) (*model.Person, error) {
	row := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT %s FROM people WHERE id = ? AND archived = 0", personColumns), id)
	p, err := scanPerson(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("person %d: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find person %d: %w", id, err)
	}
	return p, nil
}

// FindAll lists people with optional filters.
func (r *PersonRepo) FindAll(ctx context.Context, filters model.PersonFilters) ([]*model.Person, error) {
	query := fmt.Sprintf("SELECT %s FROM people WHERE archived = 0", personColumns)
	var args []any

	if filters.Tag != nil {
		query = fmt.Sprintf(
			"SELECT p.%s FROM people p JOIN taggings tg ON tg.entity_type = 'person' AND tg.entity_id = p.id JOIN tags t ON t.id = tg.tag_id WHERE p.archived = 0 AND t.name = ?",
			strings.ReplaceAll(personColumns, ", ", ", p."),
		)
		args = append(args, *filters.Tag)
	}

	if filters.OrgID != nil {
		if filters.Tag != nil {
			query += " AND p.org_id = ?"
		} else {
			query += " AND org_id = ?"
		}
		args = append(args, *filters.OrgID)
	}

	query += " ORDER BY updated_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list people: %w", err)
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

// Update modifies a person's fields.
func (r *PersonRepo) Update(ctx context.Context, id int64, input model.UpdatePersonInput) (*model.Person, error) {
	var setClauses []string
	var args []any

	if input.FirstName != nil {
		setClauses = append(setClauses, "first_name = ?")
		args = append(args, *input.FirstName)
	}
	if input.LastName != nil {
		setClauses = append(setClauses, "last_name = ?")
		args = append(args, *input.LastName)
	}
	if input.Email != nil {
		setClauses = append(setClauses, "email = ?")
		args = append(args, *input.Email)
	}
	if input.Phone != nil {
		setClauses = append(setClauses, "phone = ?")
		args = append(args, *input.Phone)
	}
	if input.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *input.Title)
	}
	if input.Company != nil {
		setClauses = append(setClauses, "company = ?")
		args = append(args, *input.Company)
	}
	if input.Location != nil {
		setClauses = append(setClauses, "location = ?")
		args = append(args, *input.Location)
	}
	if input.Notes != nil {
		setClauses = append(setClauses, "notes = ?")
		args = append(args, *input.Notes)
	}
	if input.Summary != nil {
		setClauses = append(setClauses, "summary = ?")
		args = append(args, *input.Summary)
	}
	if input.OrgID != nil {
		setClauses = append(setClauses, "org_id = ?")
		args = append(args, *input.OrgID)
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = datetime('now')")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE people SET %s WHERE id = ? AND archived = 0", strings.Join(setClauses, ", "))
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update person %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("person %d: %w", id, model.ErrNotFound)
	}

	return r.FindByID(ctx, id)
}

// Archive soft-deletes a person.
func (r *PersonRepo) Archive(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE people SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0", id)
	if err != nil {
		return fmt.Errorf("archive person %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("person %d: %w", id, model.ErrNotFound)
	}

	return nil
}

// Count returns the number of non-archived people.
func (r *PersonRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM people WHERE archived = 0").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count people: %w", err)
	}
	return count, nil
}

// Search performs full-text search on people.
func (r *PersonRepo) Search(ctx context.Context, query string, limit int) ([]*model.Person, error) {
	limit = defaultLimit(limit)

	sql := fmt.Sprintf(
		`SELECT %s FROM people WHERE archived = 0 AND id IN (SELECT rowid FROM people_fts WHERE people_fts MATCH ?) ORDER BY updated_at DESC LIMIT ?`,
		personColumns,
	)

	rows, err := r.db.QueryContext(ctx, sql, escapeFTS(query), limit)
	if err != nil {
		return nil, fmt.Errorf("search people: %w", err)
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
		return nil, fmt.Errorf("iterate search results: %w", err)
	}
	return people, nil
}
