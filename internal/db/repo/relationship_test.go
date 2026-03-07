package repo_test

import (
	"context"
	"testing"

	"github.com/jdanielnd/crm-cli/internal/db"
	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRelTestDB(t *testing.T) (*repo.RelationshipRepo, *repo.PersonRepo) {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return repo.NewRelationshipRepo(d), repo.NewPersonRepo(d)
}

func TestRelationshipCreate(t *testing.T) {
	rr, pr := setupRelTestDB(t)
	ctx := context.Background()

	p1, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	p2, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})

	rel, err := rr.Create(ctx, p1.ID, p2.ID, "colleague", nil)
	require.NoError(t, err)
	assert.Equal(t, p1.ID, rel.PersonID)
	assert.Equal(t, p2.ID, rel.RelatedPersonID)
	assert.Equal(t, "colleague", rel.Type)
}

func TestRelationshipCreate_WithNotes(t *testing.T) {
	rr, pr := setupRelTestDB(t)
	ctx := context.Background()

	p1, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	p2, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})

	notes := "Met at conference"
	rel, err := rr.Create(ctx, p1.ID, p2.ID, "friend", &notes)
	require.NoError(t, err)
	assert.Equal(t, &notes, rel.Notes)
}

func TestRelationshipCreate_InvalidType(t *testing.T) {
	rr, pr := setupRelTestDB(t)
	ctx := context.Background()

	p1, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	p2, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})

	_, err := rr.Create(ctx, p1.ID, p2.ID, "enemy", nil)
	assert.ErrorIs(t, err, model.ErrValidation)
}

func TestRelationshipCreate_SelfRelation(t *testing.T) {
	rr, pr := setupRelTestDB(t)
	ctx := context.Background()

	p1, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})

	_, err := rr.Create(ctx, p1.ID, p1.ID, "colleague", nil)
	assert.ErrorIs(t, err, model.ErrValidation)
}

func TestRelationshipFindForPerson(t *testing.T) {
	rr, pr := setupRelTestDB(t)
	ctx := context.Background()

	p1, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	p2, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})
	p3, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Alice"})

	_, _ = rr.Create(ctx, p1.ID, p2.ID, "colleague", nil)
	_, _ = rr.Create(ctx, p3.ID, p1.ID, "mentor", nil) // p1 is the related person

	rels, err := rr.FindForPerson(ctx, p1.ID)
	require.NoError(t, err)
	assert.Len(t, rels, 2) // both directions
}

func TestRelationshipDelete(t *testing.T) {
	rr, pr := setupRelTestDB(t)
	ctx := context.Background()

	p1, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	p2, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})

	rel, _ := rr.Create(ctx, p1.ID, p2.ID, "colleague", nil)

	err := rr.Delete(ctx, rel.ID)
	require.NoError(t, err)

	rels, _ := rr.FindForPerson(ctx, p1.ID)
	assert.Len(t, rels, 0)
}

func TestRelationshipDelete_NotFound(t *testing.T) {
	rr, _ := setupRelTestDB(t)
	err := rr.Delete(context.Background(), 999)
	assert.ErrorIs(t, err, model.ErrNotFound)
}
