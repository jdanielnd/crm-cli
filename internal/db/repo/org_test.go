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

func setupOrgTestDB(t *testing.T) (*repo.OrgRepo, *repo.PersonRepo) {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return repo.NewOrgRepo(d), repo.NewPersonRepo(d)
}

func TestOrgCreate(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	ctx := context.Background()

	o, err := r.Create(ctx, model.CreateOrgInput{
		Name:   "Acme Corp",
		Domain: ptr("acme.com"),
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), o.ID)
	assert.Equal(t, "Acme Corp", o.Name)
	assert.Equal(t, "acme.com", *o.Domain)
	assert.NotEmpty(t, o.UUID)
}

func TestOrgFindByID(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	ctx := context.Background()

	created, _ := r.Create(ctx, model.CreateOrgInput{Name: "Acme Corp"})
	found, err := r.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", found.Name)
}

func TestOrgFindByID_NotFound(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	_, err := r.FindByID(context.Background(), 999)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestOrgFindAll(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	ctx := context.Background()

	r.Create(ctx, model.CreateOrgInput{Name: "Acme"})
	r.Create(ctx, model.CreateOrgInput{Name: "Globex"})

	orgs, err := r.FindAll(ctx, 0)
	require.NoError(t, err)
	assert.Len(t, orgs, 2)
}

func TestOrgFindAll_WithLimit(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	ctx := context.Background()

	r.Create(ctx, model.CreateOrgInput{Name: "Acme"})
	r.Create(ctx, model.CreateOrgInput{Name: "Globex"})
	r.Create(ctx, model.CreateOrgInput{Name: "Initech"})

	orgs, err := r.FindAll(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, orgs, 2)
}

func TestOrgUpdate(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	ctx := context.Background()

	r.Create(ctx, model.CreateOrgInput{Name: "Acme"})

	newDomain := "acme.io"
	updated, err := r.Update(ctx, 1, model.UpdateOrgInput{Domain: &newDomain})
	require.NoError(t, err)
	assert.Equal(t, "acme.io", *updated.Domain)
	assert.Equal(t, "Acme", updated.Name)
}

func TestOrgUpdate_NotFound(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	name := "Test"
	_, err := r.Update(context.Background(), 999, model.UpdateOrgInput{Name: &name})
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestOrgArchive(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	ctx := context.Background()

	r.Create(ctx, model.CreateOrgInput{Name: "Acme"})
	err := r.Archive(ctx, 1)
	require.NoError(t, err)

	_, err = r.FindByID(ctx, 1)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestOrgArchive_NotFound(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	err := r.Archive(context.Background(), 999)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestOrgSearch(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	ctx := context.Background()

	r.Create(ctx, model.CreateOrgInput{Name: "Acme Corp", Domain: ptr("acme.com")})
	r.Create(ctx, model.CreateOrgInput{Name: "Globex Inc"})

	results, err := r.Search(ctx, "Acme", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Acme Corp", results[0].Name)
}

func TestOrgSearch_ArchivedExcluded(t *testing.T) {
	r, _ := setupOrgTestDB(t)
	ctx := context.Background()

	r.Create(ctx, model.CreateOrgInput{Name: "Acme Corp"})
	r.Archive(ctx, 1)

	results, err := r.Search(ctx, "Acme", 10)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestOrgFindPeople(t *testing.T) {
	orgRepo, personRepo := setupOrgTestDB(t)
	ctx := context.Background()

	org, _ := orgRepo.Create(ctx, model.CreateOrgInput{Name: "Acme"})

	personRepo.Create(ctx, model.CreatePersonInput{FirstName: "Jane", OrgID: &org.ID})
	personRepo.Create(ctx, model.CreatePersonInput{FirstName: "Bob", OrgID: &org.ID})
	personRepo.Create(ctx, model.CreatePersonInput{FirstName: "Alice"}) // no org

	people, err := orgRepo.FindPeople(ctx, org.ID)
	require.NoError(t, err)
	assert.Len(t, people, 2)
}
