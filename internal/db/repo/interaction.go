package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jdanielnd/crm-cli/internal/model"
)

// InteractionRepo handles interaction database operations.
type InteractionRepo struct {
	db *sql.DB
}

// NewInteractionRepo creates a new InteractionRepo.
func NewInteractionRepo(db *sql.DB) *InteractionRepo {
	return &InteractionRepo{db: db}
}

var interactionColumns = "id, uuid, type, subject, content, direction, occurred_at, created_at, updated_at"

func scanInteraction(row interface{ Scan(...any) error }) (*model.Interaction, error) {
	var i model.Interaction
	err := row.Scan(
		&i.ID, &i.UUID, &i.Type, &i.Subject, &i.Content,
		&i.Direction, &i.OccurredAt, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// Create inserts a new interaction and links it to people.
func (r *InteractionRepo) Create(ctx context.Context, input model.CreateInteractionInput) (*model.Interaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id := uuid.New().String()
	occurredAt := "datetime('now')"
	args := []any{id, input.Type, input.Subject, input.Content, input.Direction}

	var query string
	if input.OccurredAt != nil {
		query = `INSERT INTO interactions (uuid, type, subject, content, direction, occurred_at)
				 VALUES (?, ?, ?, ?, ?, ?)`
		args = append(args, *input.OccurredAt)
	} else {
		query = fmt.Sprintf(`INSERT INTO interactions (uuid, type, subject, content, direction, occurred_at)
				 VALUES (?, ?, ?, ?, ?, %s)`, occurredAt)
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("insert interaction: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	// Link to people
	for _, personID := range input.PersonIDs {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO interaction_people (interaction_id, person_id) VALUES (?, ?)",
			rowID, personID)
		if err != nil {
			return nil, fmt.Errorf("link interaction %d to person %d: %w", rowID, personID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return r.FindByID(ctx, rowID)
}

// FindByID returns an interaction by ID with its person IDs.
func (r *InteractionRepo) FindByID(ctx context.Context, id int64) (*model.Interaction, error) {
	row := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT %s FROM interactions WHERE id = ? AND archived = 0", interactionColumns), id)
	i, err := scanInteraction(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("interaction %d: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find interaction %d: %w", id, err)
	}

	personIDs, err := r.getPersonIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	i.PersonIDs = personIDs

	return i, nil
}

// FindAll lists interactions with optional filters.
func (r *InteractionRepo) FindAll(ctx context.Context, filters model.InteractionFilters) ([]*model.Interaction, error) {
	query := fmt.Sprintf("SELECT %s FROM interactions WHERE archived = 0", interactionColumns)
	var args []any

	if filters.PersonID != nil {
		query = fmt.Sprintf(
			"SELECT i.%s FROM interactions i JOIN interaction_people ip ON ip.interaction_id = i.id WHERE i.archived = 0 AND ip.person_id = ?",
			strings.ReplaceAll(interactionColumns, ", ", ", i."),
		)
		args = append(args, *filters.PersonID)
	}

	if filters.Type != nil {
		if filters.PersonID != nil {
			query += " AND i.type = ?"
		} else {
			query += " AND type = ?"
		}
		args = append(args, *filters.Type)
	}

	query += " ORDER BY occurred_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list interactions: %w", err)
	}
	defer rows.Close()

	// Collect interactions first, then close rows before querying person IDs.
	// This avoids deadlock with MaxOpenConns(1) since getPersonIDs needs
	// its own connection.
	var interactions []*model.Interaction
	for rows.Next() {
		i, err := scanInteraction(rows)
		if err != nil {
			return nil, fmt.Errorf("scan interaction: %w", err)
		}
		interactions = append(interactions, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate interactions: %w", err)
	}
	rows.Close()

	// Now fetch person IDs for each interaction
	for _, i := range interactions {
		personIDs, err := r.getPersonIDs(ctx, i.ID)
		if err != nil {
			return nil, err
		}
		i.PersonIDs = personIDs
	}

	return interactions, nil
}

// Search finds interactions matching a full-text query.
func (r *InteractionRepo) Search(ctx context.Context, query string, limit int) ([]*model.Interaction, error) {
	if limit <= 0 {
		limit = 20
	}

	ftsQuery := `"` + strings.ReplaceAll(query, `"`, `""`) + `"` + "*"
	sqlStr := fmt.Sprintf(
		`SELECT %s FROM interactions WHERE archived = 0 AND id IN (SELECT rowid FROM interactions_fts WHERE interactions_fts MATCH ?) ORDER BY occurred_at DESC LIMIT ?`,
		interactionColumns,
	)

	rows, err := r.db.QueryContext(ctx, sqlStr, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search interactions: %w", err)
	}
	defer rows.Close()

	var interactions []*model.Interaction
	for rows.Next() {
		i, err := scanInteraction(rows)
		if err != nil {
			return nil, fmt.Errorf("scan interaction: %w", err)
		}
		interactions = append(interactions, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate interactions: %w", err)
	}
	rows.Close()

	for _, i := range interactions {
		personIDs, err := r.getPersonIDs(ctx, i.ID)
		if err != nil {
			return nil, err
		}
		i.PersonIDs = personIDs
	}

	return interactions, nil
}

func (r *InteractionRepo) getPersonIDs(ctx context.Context, interactionID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT person_id FROM interaction_people WHERE interaction_id = ?", interactionID)
	if err != nil {
		return nil, fmt.Errorf("get person IDs for interaction %d: %w", interactionID, err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan person ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
