package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jdanielnd/crm-cli/internal/model"
)

// RelationshipRepo handles relationship database operations.
type RelationshipRepo struct {
	db *sql.DB
}

// NewRelationshipRepo creates a new RelationshipRepo.
func NewRelationshipRepo(db *sql.DB) *RelationshipRepo {
	return &RelationshipRepo{db: db}
}

// Create adds a new relationship between two people.
func (r *RelationshipRepo) Create(ctx context.Context, personID, relatedID int64, relType string, notes *string) (*model.Relationship, error) {
	if !model.ValidRelationshipType(relType) {
		return nil, fmt.Errorf("invalid relationship type %q: %w", relType, model.ErrValidation)
	}
	if personID == relatedID {
		return nil, fmt.Errorf("cannot relate a person to themselves: %w", model.ErrValidation)
	}

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO relationships (person_id, related_person_id, type, notes) VALUES (?, ?, ?, ?)",
		personID, relatedID, relType, notes)
	if err != nil {
		return nil, fmt.Errorf("create relationship: %w", err)
	}

	id, _ := result.LastInsertId()
	return &model.Relationship{
		ID:              id,
		PersonID:        personID,
		RelatedPersonID: relatedID,
		Type:            relType,
		Notes:           notes,
	}, nil
}

// FindForPerson returns all relationships for a person (in both directions).
func (r *RelationshipRepo) FindForPerson(ctx context.Context, personID int64) ([]*model.Relationship, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, person_id, related_person_id, type, notes, created_at
		 FROM relationships
		 WHERE person_id = ? OR related_person_id = ?
		 ORDER BY created_at DESC`,
		personID, personID)
	if err != nil {
		return nil, fmt.Errorf("find relationships for person %d: %w", personID, err)
	}
	defer rows.Close()

	var rels []*model.Relationship
	for rows.Next() {
		var rel model.Relationship
		if err := rows.Scan(&rel.ID, &rel.PersonID, &rel.RelatedPersonID, &rel.Type, &rel.Notes, &rel.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan relationship: %w", err)
		}
		rels = append(rels, &rel)
	}
	return rels, rows.Err()
}

// Delete removes a relationship by ID.
func (r *RelationshipRepo) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM relationships WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete relationship: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("relationship %d: %w", id, model.ErrNotFound)
	}
	return nil
}
