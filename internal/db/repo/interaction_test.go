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

func setupInteractionTestDB(t *testing.T) (*repo.InteractionRepo, *repo.PersonRepo) {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return repo.NewInteractionRepo(d), repo.NewPersonRepo(d)
}

func TestInteractionCreate(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)

	i, err := ir.Create(ctx, model.CreateInteractionInput{
		Type:      "call",
		Subject:   ptr("Intro call"),
		PersonIDs: []int64{p.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, "call", i.Type)
	assert.Equal(t, "Intro call", *i.Subject)
	assert.Equal(t, []int64{p.ID}, i.PersonIDs)
}

func TestInteractionCreate_MultiplePeople(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p1, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)
	p2, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})
	require.NoError(t, err)

	i, err := ir.Create(ctx, model.CreateInteractionInput{
		Type:      "meeting",
		Subject:   ptr("Team sync"),
		PersonIDs: []int64{p1.ID, p2.ID},
	})
	require.NoError(t, err)
	assert.Len(t, i.PersonIDs, 2)
}

func TestInteractionCreate_WithDirection(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)

	i, err := ir.Create(ctx, model.CreateInteractionInput{
		Type:      "email",
		Subject:   ptr("Follow-up"),
		Direction: ptr("outbound"),
		PersonIDs: []int64{p.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, "outbound", *i.Direction)
}

func TestInteractionCreate_WithCustomDate(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)
	at := "2026-03-05 14:00:00"

	i, err := ir.Create(ctx, model.CreateInteractionInput{
		Type:       "meeting",
		Subject:    ptr("Demo"),
		OccurredAt: &at,
		PersonIDs:  []int64{p.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, "2026-03-05 14:00:00", i.OccurredAt)
}

func TestInteractionFindByID(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)
	created, err := ir.Create(ctx, model.CreateInteractionInput{
		Type:      "call",
		PersonIDs: []int64{p.ID},
	})
	require.NoError(t, err)

	found, err := ir.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, []int64{p.ID}, found.PersonIDs)
}

func TestInteractionFindByID_NotFound(t *testing.T) {
	ir, _ := setupInteractionTestDB(t)
	_, err := ir.FindByID(context.Background(), 999)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestInteractionFindAll(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "call", PersonIDs: []int64{p.ID}})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "email", PersonIDs: []int64{p.ID}})
	require.NoError(t, err)

	results, err := ir.FindAll(ctx, model.InteractionFilters{})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestInteractionFindAll_FilterByPerson(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p1, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)
	p2, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Bob"})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "call", PersonIDs: []int64{p1.ID}})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "email", PersonIDs: []int64{p2.ID}})
	require.NoError(t, err)

	results, err := ir.FindAll(ctx, model.InteractionFilters{PersonID: &p1.ID})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "call", results[0].Type)
}

func TestInteractionFindAll_FilterByType(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "call", PersonIDs: []int64{p.ID}})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "email", PersonIDs: []int64{p.ID}})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "call", PersonIDs: []int64{p.ID}})
	require.NoError(t, err)

	callType := "call"
	results, err := ir.FindAll(ctx, model.InteractionFilters{Type: &callType})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestInteractionFindAll_WithLimit(t *testing.T) {
	ir, pr := setupInteractionTestDB(t)
	ctx := context.Background()

	p, err := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "call", PersonIDs: []int64{p.ID}})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "email", PersonIDs: []int64{p.ID}})
	require.NoError(t, err)
	_, err = ir.Create(ctx, model.CreateInteractionInput{Type: "note", PersonIDs: []int64{p.ID}})
	require.NoError(t, err)

	results, err := ir.FindAll(ctx, model.InteractionFilters{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}
