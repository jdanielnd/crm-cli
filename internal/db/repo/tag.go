package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jdanielnd/crm-cli/internal/model"
)

// TagRepo handles tag database operations.
type TagRepo struct {
	db *sql.DB
}

// NewTagRepo creates a new TagRepo.
func NewTagRepo(db *sql.DB) *TagRepo {
	return &TagRepo{db: db}
}

// FindAll lists all tags.
func (r *TagRepo) FindAll(ctx context.Context) ([]*model.Tag, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM tags ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, &t)
	}
	return tags, rows.Err()
}

// FindOrCreate returns a tag by name, creating it if it doesn't exist.
func (r *TagRepo) FindOrCreate(ctx context.Context, name string) (*model.Tag, error) {
	var t model.Tag
	err := r.db.QueryRowContext(ctx, "SELECT id, name FROM tags WHERE name = ?", name).Scan(&t.ID, &t.Name)
	if err == nil {
		return &t, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find tag: %w", err)
	}

	result, err := r.db.ExecContext(ctx, "INSERT INTO tags (name) VALUES (?)", name)
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	id, _ := result.LastInsertId()
	return &model.Tag{ID: id, Name: name}, nil
}

// Apply adds a tag to an entity. No-op if already applied.
func (r *TagRepo) Apply(ctx context.Context, entityType string, entityID int64, tagName string) error {
	if !model.ValidEntityType(entityType) {
		return fmt.Errorf("invalid entity type %q: %w", entityType, model.ErrValidation)
	}

	// Verify entity exists. Table name is safe — entityType is validated above.
	tableMap := map[string]string{
		"person": "people", "organization": "organizations",
		"deal": "deals", "interaction": "interactions",
	}
	table := tableMap[entityType]
	var exists int
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT 1 FROM %s WHERE id = ? AND archived = 0", table), entityID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s %d: %w", entityType, entityID, model.ErrNotFound)
	}

	tag, err := r.FindOrCreate(ctx, tagName)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO taggings (tag_id, entity_type, entity_id) VALUES (?, ?, ?)",
		tag.ID, entityType, entityID)
	if err != nil {
		return fmt.Errorf("apply tag: %w", err)
	}
	return nil
}

// Remove removes a tag from an entity.
func (r *TagRepo) Remove(ctx context.Context, entityType string, entityID int64, tagName string) error {
	if !model.ValidEntityType(entityType) {
		return fmt.Errorf("invalid entity type %q: %w", entityType, model.ErrValidation)
	}

	result, err := r.db.ExecContext(ctx,
		`DELETE FROM taggings WHERE tag_id = (SELECT id FROM tags WHERE name = ?) AND entity_type = ? AND entity_id = ?`,
		tagName, entityType, entityID)
	if err != nil {
		return fmt.Errorf("remove tag: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tag %q on %s %d: %w", tagName, entityType, entityID, model.ErrNotFound)
	}
	return nil
}

// GetForEntity returns all tags for an entity.
func (r *TagRepo) GetForEntity(ctx context.Context, entityType string, entityID int64) ([]*model.Tag, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.id, t.name FROM tags t JOIN taggings tg ON tg.tag_id = t.id
		 WHERE tg.entity_type = ? AND tg.entity_id = ? ORDER BY t.name`,
		entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("get tags for %s %d: %w", entityType, entityID, err)
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, &t)
	}
	return tags, rows.Err()
}

// GetEntities returns entity IDs with a given tag.
func (r *TagRepo) GetEntities(ctx context.Context, tagName string, entityType string) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT tg.entity_id FROM taggings tg JOIN tags t ON t.id = tg.tag_id
		 WHERE t.name = ? AND tg.entity_type = ?`,
		tagName, entityType)
	if err != nil {
		return nil, fmt.Errorf("get entities for tag %q: %w", tagName, err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan entity id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Delete removes a tag and all its taggings.
func (r *TagRepo) Delete(ctx context.Context, tagName string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tags WHERE name = ?", tagName)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tag %q: %w", tagName, model.ErrNotFound)
	}
	return nil
}
