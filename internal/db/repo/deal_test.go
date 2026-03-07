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

func setupDealTestDB(t *testing.T) (*repo.DealRepo, *repo.PersonRepo) {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return repo.NewDealRepo(d), repo.NewPersonRepo(d)
}

func TestDealCreate(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	ctx := context.Background()

	deal, err := dr.Create(ctx, model.CreateDealInput{
		Title: "Website Redesign",
		Stage: "lead",
	})
	require.NoError(t, err)
	assert.Equal(t, "Website Redesign", deal.Title)
	assert.Equal(t, "lead", deal.Stage)
	assert.NotEmpty(t, deal.UUID)
}

func TestDealCreate_WithValue(t *testing.T) {
	dr, pr := setupDealTestDB(t)
	ctx := context.Background()

	p, _ := pr.Create(ctx, model.CreatePersonInput{FirstName: "Jane"})
	value := 15000.0

	deal, err := dr.Create(ctx, model.CreateDealInput{
		Title:    "Consulting",
		Stage:    "proposal",
		Value:    &value,
		PersonID: &p.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, &value, deal.Value)
	assert.Equal(t, &p.ID, deal.PersonID)
}

func TestDealCreate_InvalidStage(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	_, err := dr.Create(context.Background(), model.CreateDealInput{
		Title: "Bad Deal",
		Stage: "invalid",
	})
	assert.ErrorIs(t, err, model.ErrValidation)
}

func TestDealFindByID(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	ctx := context.Background()

	created, _ := dr.Create(ctx, model.CreateDealInput{Title: "Test", Stage: "lead"})

	found, err := dr.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestDealFindByID_NotFound(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	_, err := dr.FindByID(context.Background(), 999)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestDealFindAll(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	ctx := context.Background()

	_, _ = dr.Create(ctx, model.CreateDealInput{Title: "Deal 1", Stage: "lead"})
	_, _ = dr.Create(ctx, model.CreateDealInput{Title: "Deal 2", Stage: "proposal"})

	deals, err := dr.FindAll(ctx, model.DealFilters{})
	require.NoError(t, err)
	assert.Len(t, deals, 2)
}

func TestDealFindAll_FilterByStage(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	ctx := context.Background()

	_, _ = dr.Create(ctx, model.CreateDealInput{Title: "Deal 1", Stage: "lead"})
	_, _ = dr.Create(ctx, model.CreateDealInput{Title: "Deal 2", Stage: "proposal"})

	stage := "proposal"
	deals, err := dr.FindAll(ctx, model.DealFilters{Stage: &stage})
	require.NoError(t, err)
	assert.Len(t, deals, 1)
	assert.Equal(t, "Deal 2", deals[0].Title)
}

func TestDealUpdate(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	ctx := context.Background()

	deal, _ := dr.Create(ctx, model.CreateDealInput{Title: "Test", Stage: "lead"})

	newStage := "proposal"
	value := 5000.0
	updated, err := dr.Update(ctx, deal.ID, model.UpdateDealInput{
		Stage: &newStage,
		Value: &value,
	})
	require.NoError(t, err)
	assert.Equal(t, "proposal", updated.Stage)
	assert.Equal(t, &value, updated.Value)
}

func TestDealUpdate_InvalidStage(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	ctx := context.Background()

	deal, _ := dr.Create(ctx, model.CreateDealInput{Title: "Test", Stage: "lead"})

	badStage := "invalid"
	_, err := dr.Update(ctx, deal.ID, model.UpdateDealInput{Stage: &badStage})
	assert.ErrorIs(t, err, model.ErrValidation)
}

func TestDealArchive(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	ctx := context.Background()

	deal, _ := dr.Create(ctx, model.CreateDealInput{Title: "Test", Stage: "lead"})

	err := dr.Archive(ctx, deal.ID)
	require.NoError(t, err)

	_, err = dr.FindByID(ctx, deal.ID)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestDealPipeline(t *testing.T) {
	dr, _ := setupDealTestDB(t)
	ctx := context.Background()

	v1 := 10000.0
	v2 := 5000.0
	_, _ = dr.Create(ctx, model.CreateDealInput{Title: "Deal 1", Stage: "lead", Value: &v1})
	_, _ = dr.Create(ctx, model.CreateDealInput{Title: "Deal 2", Stage: "lead", Value: &v2})
	_, _ = dr.Create(ctx, model.CreateDealInput{Title: "Deal 3", Stage: "won", Value: &v1})

	stages, err := dr.Pipeline(ctx)
	require.NoError(t, err)
	assert.Len(t, stages, 2)
	assert.Equal(t, "lead", stages[0].Stage)
	assert.Equal(t, 2, stages[0].Count)
	assert.Equal(t, 15000.0, stages[0].TotalValue)
}
