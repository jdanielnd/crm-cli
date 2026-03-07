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

func setupTestDB(t *testing.T) *repo.PersonRepo {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return repo.NewPersonRepo(d)
}

func ptr(s string) *string { return &s }

func TestPersonCreate(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	p, err := r.Create(ctx, model.CreatePersonInput{
		FirstName: "Jane",
		LastName:  ptr("Smith"),
		Email:     ptr("jane@example.com"),
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
	assert.Equal(t, "Jane", p.FirstName)
	assert.Equal(t, "Smith", *p.LastName)
	assert.Equal(t, "jane@example.com", *p.Email)
	assert.NotEmpty(t, p.UUID)
	assert.NotEmpty(t, p.CreatedAt)
}

func TestPersonFindByID(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	created, _ := r.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	found, err := r.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Jane", found.FirstName)
}

func TestPersonFindByID_NotFound(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, err := r.FindByID(ctx, 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestPersonFindAll(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})

	people, err := r.FindAll(ctx, model.PersonFilters{})
	require.NoError(t, err)
	assert.Len(t, people, 2)
}

func TestPersonFindAll_WithLimit(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})
	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Alice"})

	people, err := r.FindAll(ctx, model.PersonFilters{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, people, 2)
}

func TestPersonUpdate(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})

	newEmail := "new@example.com"
	updated, err := r.Update(ctx, 1, model.UpdatePersonInput{Email: &newEmail})
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", *updated.Email)
	assert.Equal(t, "Jane", updated.FirstName)
}

func TestPersonUpdate_NotFound(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	name := "Test"
	_, err := r.Update(ctx, 999, model.UpdatePersonInput{FirstName: &name})
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestPersonUpdate_NoChanges(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})

	// Empty update should return the person unchanged
	p, err := r.Update(ctx, 1, model.UpdatePersonInput{})
	require.NoError(t, err)
	assert.Equal(t, "Jane", p.FirstName)
}

func TestPersonArchive(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	err := r.Archive(ctx, 1)
	require.NoError(t, err)

	// Should not be found
	_, err = r.FindByID(ctx, 1)
	assert.ErrorIs(t, err, model.ErrNotFound)

	// Should not appear in list
	people, _ := r.FindAll(ctx, model.PersonFilters{})
	assert.Len(t, people, 0)
}

func TestPersonArchive_NotFound(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	err := r.Archive(ctx, 999)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestPersonSearch(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane", LastName: ptr("Smith"), Email: ptr("jane@example.com")})
	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Bob", LastName: ptr("Jones")})

	results, err := r.Search(ctx, "Jane", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Jane", results[0].FirstName)
}

func TestPersonSearch_ByEmail(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane", Email: ptr("jane@example.com")})

	results, err := r.Search(ctx, "jane@example", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestPersonSearch_NoResults(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})

	results, err := r.Search(ctx, "nonexistent", 10)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestPersonCreate_DuplicateEmail(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, err := r.Create(ctx, model.CreatePersonInput{
		FirstName: "Jane",
		Email:     ptr("jane@example.com"),
	})
	require.NoError(t, err)

	_, err = r.Create(ctx, model.CreatePersonInput{
		FirstName: "Janet",
		Email:     ptr("jane@example.com"),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrConflict)
	assert.Contains(t, err.Error(), "jane@example.com")
}

func TestPersonCreate_DuplicateName(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, err := r.Create(ctx, model.CreatePersonInput{
		FirstName: "Jane",
		LastName:  ptr("Smith"),
	})
	require.NoError(t, err)

	_, err = r.Create(ctx, model.CreatePersonInput{
		FirstName: "Jane",
		LastName:  ptr("Smith"),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrConflict)
	assert.Contains(t, err.Error(), "Jane")
}

func TestPersonCreate_DuplicateAllowedAfterArchive(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, err := r.Create(ctx, model.CreatePersonInput{
		FirstName: "Jane",
		LastName:  ptr("Smith"),
		Email:     ptr("jane@example.com"),
	})
	require.NoError(t, err)

	err = r.Archive(ctx, 1)
	require.NoError(t, err)

	// Should succeed since original is archived
	_, err = r.Create(ctx, model.CreatePersonInput{
		FirstName: "Jane",
		LastName:  ptr("Smith"),
		Email:     ptr("jane@example.com"),
	})
	require.NoError(t, err)
}

func TestPersonSearch_ArchivedExcluded(t *testing.T) {
	r := setupTestDB(t)
	ctx := context.Background()

	_, _ = r.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	_ = r.Archive(ctx, 1)

	results, err := r.Search(ctx, "Jane", 10)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}
