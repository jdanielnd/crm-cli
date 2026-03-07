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

func setupTagTestDB(t *testing.T) (*repo.TagRepo, *repo.PersonRepo) {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return repo.NewTagRepo(d), repo.NewPersonRepo(d)
}

func TestTagFindOrCreate(t *testing.T) {
	r, _ := setupTagTestDB(t)
	ctx := context.Background()

	tag, err := r.FindOrCreate(ctx, "vip")
	require.NoError(t, err)
	assert.Equal(t, "vip", tag.Name)

	// Second call should return the same tag
	tag2, err := r.FindOrCreate(ctx, "vip")
	require.NoError(t, err)
	assert.Equal(t, tag.ID, tag2.ID)
}

func TestTagApplyAndGetForEntity(t *testing.T) {
	tr, pr := setupTagTestDB(t)
	ctx := context.Background()

	p, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})

	err := tr.Apply(ctx, "person", p.ID, "vip")
	require.NoError(t, err)
	err = tr.Apply(ctx, "person", p.ID, "client")
	require.NoError(t, err)

	tags, err := tr.GetForEntity(ctx, "person", p.ID)
	require.NoError(t, err)
	assert.Len(t, tags, 2)
}

func TestTagApply_Idempotent(t *testing.T) {
	tr, pr := setupTagTestDB(t)
	ctx := context.Background()

	p, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})

	_ = tr.Apply(ctx, "person", p.ID, "vip")
	err := tr.Apply(ctx, "person", p.ID, "vip") // duplicate
	require.NoError(t, err)

	tags, _ := tr.GetForEntity(ctx, "person", p.ID)
	assert.Len(t, tags, 1)
}

func TestTagApply_InvalidEntityType(t *testing.T) {
	tr, _ := setupTagTestDB(t)
	err := tr.Apply(context.Background(), "invalid", 1, "vip")
	assert.ErrorIs(t, err, model.ErrValidation)
}

func TestTagRemove(t *testing.T) {
	tr, pr := setupTagTestDB(t)
	ctx := context.Background()

	p, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	_ = tr.Apply(ctx, "person", p.ID, "vip")

	err := tr.Remove(ctx, "person", p.ID, "vip")
	require.NoError(t, err)

	tags, _ := tr.GetForEntity(ctx, "person", p.ID)
	assert.Len(t, tags, 0)
}

func TestTagRemove_NotFound(t *testing.T) {
	tr, _ := setupTagTestDB(t)
	err := tr.Remove(context.Background(), "person", 1, "nonexistent")
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestTagFindAll(t *testing.T) {
	tr, pr := setupTagTestDB(t)
	ctx := context.Background()

	p, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	_ = tr.Apply(ctx, "person", p.ID, "vip")
	_ = tr.Apply(ctx, "person", p.ID, "client")

	tags, err := tr.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, tags, 2)
}

func TestTagGetEntities(t *testing.T) {
	tr, pr := setupTagTestDB(t)
	ctx := context.Background()

	p1, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	p2, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})
	_ = tr.Apply(ctx, "person", p1.ID, "vip")
	_ = tr.Apply(ctx, "person", p2.ID, "vip")

	ids, err := tr.GetEntities(ctx, "vip", "person")
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestTagDelete(t *testing.T) {
	tr, pr := setupTagTestDB(t)
	ctx := context.Background()

	p, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	_ = tr.Apply(ctx, "person", p.ID, "vip")

	err := tr.Delete(ctx, "vip")
	require.NoError(t, err)

	tags, _ := tr.FindAll(ctx)
	assert.Len(t, tags, 0)

	// Taggings should be gone too (CASCADE)
	entityTags, _ := tr.GetForEntity(ctx, "person", p.ID)
	assert.Len(t, entityTags, 0)
}

func TestTagDelete_NotFound(t *testing.T) {
	tr, _ := setupTagTestDB(t)
	err := tr.Delete(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, model.ErrNotFound)
}
