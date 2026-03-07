package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jdanielnd/crm-cli/internal/model"
)

// DealRepo handles deal database operations.
type DealRepo struct {
	db *sql.DB
}

// NewDealRepo creates a new DealRepo.
func NewDealRepo(db *sql.DB) *DealRepo {
	return &DealRepo{db: db}
}

func scanDeal(row interface{ Scan(...any) error }) (*model.Deal, error) {
	var d model.Deal
	err := row.Scan(
		&d.ID, &d.UUID, &d.Title, &d.Value, &d.Stage,
		&d.PersonID, &d.OrgID, &d.Notes, &d.ClosedAt,
		&d.Archived, &d.CreatedAt, &d.UpdatedAt,
	)
	return &d, err
}

// Create inserts a new deal.
func (r *DealRepo) Create(ctx context.Context, input model.CreateDealInput) (*model.Deal, error) {
	if !model.ValidDealStage(input.Stage) {
		return nil, fmt.Errorf("invalid deal stage %q: %w", input.Stage, model.ErrValidation)
	}

	id := uuid.New().String()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO deals (uuid, title, value, stage, person_id, org_id, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, input.Title, input.Value, input.Stage, input.PersonID, input.OrgID, input.Notes)
	if err != nil {
		return nil, fmt.Errorf("create deal: %w", err)
	}

	dealID, _ := result.LastInsertId()
	return r.FindByID(ctx, dealID)
}

// FindByID returns a deal by ID.
func (r *DealRepo) FindByID(ctx context.Context, id int64) (*model.Deal, error) {
	d, err := scanDeal(r.db.QueryRowContext(ctx,
		`SELECT id, uuid, title, value, stage, person_id, org_id, notes, closed_at,
		        archived, created_at, updated_at
		 FROM deals WHERE id = ? AND archived = 0`, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deal %d: %w", id, model.ErrNotFound)
		}
		return nil, fmt.Errorf("find deal %d: %w", id, err)
	}
	return d, nil
}

// FindAll returns deals with optional filters.
func (r *DealRepo) FindAll(ctx context.Context, filters model.DealFilters) ([]*model.Deal, error) {
	query := `SELECT id, uuid, title, value, stage, person_id, org_id, notes, closed_at,
	                 archived, created_at, updated_at
	          FROM deals WHERE archived = 0`
	var args []any

	if filters.Stage != nil {
		query += " AND stage = ?"
		args = append(args, *filters.Stage)
	}
	if filters.PersonID != nil {
		query += " AND person_id = ?"
		args = append(args, *filters.PersonID)
	}
	if filters.OrgID != nil {
		query += " AND org_id = ?"
		args = append(args, *filters.OrgID)
	}
	if filters.ExcludeClosed {
		query += " AND stage NOT IN ('won', 'lost')"
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list deals: %w", err)
	}
	defer rows.Close()

	var deals []*model.Deal
	for rows.Next() {
		d, err := scanDeal(rows)
		if err != nil {
			return nil, fmt.Errorf("scan deal: %w", err)
		}
		deals = append(deals, d)
	}
	return deals, rows.Err()
}

// Update modifies a deal.
func (r *DealRepo) Update(ctx context.Context, id int64, input model.UpdateDealInput) (*model.Deal, error) {
	var setClauses []string
	var args []any

	if input.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *input.Title)
	}
	if input.Value != nil {
		setClauses = append(setClauses, "value = ?")
		args = append(args, *input.Value)
	}
	if input.Stage != nil {
		if !model.ValidDealStage(*input.Stage) {
			return nil, fmt.Errorf("invalid deal stage %q: %w", *input.Stage, model.ErrValidation)
		}
		setClauses = append(setClauses, "stage = ?")
		args = append(args, *input.Stage)
	}
	if input.PersonID != nil {
		setClauses = append(setClauses, "person_id = ?")
		args = append(args, *input.PersonID)
	}
	if input.OrgID != nil {
		setClauses = append(setClauses, "org_id = ?")
		args = append(args, *input.OrgID)
	}
	if input.Notes != nil {
		setClauses = append(setClauses, "notes = ?")
		args = append(args, *input.Notes)
	}
	if input.ClosedAt != nil {
		setClauses = append(setClauses, "closed_at = ?")
		args = append(args, *input.ClosedAt)
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = datetime('now')")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE deals SET %s WHERE id = ? AND archived = 0", strings.Join(setClauses, ", "))
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update deal %d: %w", id, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("deal %d: %w", id, model.ErrNotFound)
	}

	return r.FindByID(ctx, id)
}

// Archive soft-deletes a deal.
func (r *DealRepo) Archive(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE deals SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0", id)
	if err != nil {
		return fmt.Errorf("archive deal %d: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("deal %d: %w", id, model.ErrNotFound)
	}
	return nil
}

// Search finds deals matching a full-text query.
func (r *DealRepo) Search(ctx context.Context, query string, limit int) ([]*model.Deal, error) {
	if limit <= 0 {
		limit = 20
	}

	ftsQuery := `"` + strings.ReplaceAll(query, `"`, `""`) + `"` + "*"
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, uuid, title, value, stage, person_id, org_id, notes, closed_at,
		        archived, created_at, updated_at
		 FROM deals WHERE archived = 0 AND id IN (SELECT rowid FROM deals_fts WHERE deals_fts MATCH ?)
		 ORDER BY updated_at DESC LIMIT ?`,
		ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search deals: %w", err)
	}
	defer rows.Close()

	var deals []*model.Deal
	for rows.Next() {
		d, err := scanDeal(rows)
		if err != nil {
			return nil, fmt.Errorf("scan deal: %w", err)
		}
		deals = append(deals, d)
	}
	return deals, rows.Err()
}

// Pipeline returns a summary of deals grouped by stage.
func (r *DealRepo) Pipeline(ctx context.Context) ([]*model.PipelineStage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT stage, COUNT(*) as count, COALESCE(SUM(value), 0) as total_value
		 FROM deals WHERE archived = 0
		 GROUP BY stage
		 ORDER BY CASE stage
		   WHEN 'lead' THEN 1
		   WHEN 'prospect' THEN 2
		   WHEN 'proposal' THEN 3
		   WHEN 'negotiation' THEN 4
		   WHEN 'won' THEN 5
		   WHEN 'lost' THEN 6
		 END`)
	if err != nil {
		return nil, fmt.Errorf("pipeline: %w", err)
	}
	defer rows.Close()

	var stages []*model.PipelineStage
	for rows.Next() {
		var s model.PipelineStage
		if err := rows.Scan(&s.Stage, &s.Count, &s.TotalValue); err != nil {
			return nil, fmt.Errorf("scan pipeline stage: %w", err)
		}
		stages = append(stages, &s)
	}
	return stages, rows.Err()
}
