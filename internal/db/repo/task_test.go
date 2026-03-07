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

func setupTaskTestDB(t *testing.T) *repo.TaskRepo {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return repo.NewTaskRepo(d)
}

func TestTaskCreate(t *testing.T) {
	tr := setupTaskTestDB(t)
	ctx := context.Background()

	task, err := tr.Create(ctx, model.CreateTaskInput{
		Title:    "Follow up",
		Priority: "high",
	})
	require.NoError(t, err)
	assert.Equal(t, "Follow up", task.Title)
	assert.Equal(t, "high", task.Priority)
	assert.False(t, task.Completed)
}

func TestTaskCreate_WithDue(t *testing.T) {
	tr := setupTaskTestDB(t)
	ctx := context.Background()

	due := "2026-03-14"
	task, err := tr.Create(ctx, model.CreateTaskInput{
		Title:    "Call back",
		Priority: "medium",
		DueAt:    &due,
	})
	require.NoError(t, err)
	assert.Equal(t, &due, task.DueAt)
}

func TestTaskCreate_InvalidPriority(t *testing.T) {
	tr := setupTaskTestDB(t)
	_, err := tr.Create(context.Background(), model.CreateTaskInput{
		Title:    "Bad",
		Priority: "urgent",
	})
	assert.ErrorIs(t, err, model.ErrValidation)
}

func TestTaskFindByID(t *testing.T) {
	tr := setupTaskTestDB(t)
	ctx := context.Background()

	created, _ := tr.Create(ctx, model.CreateTaskInput{Title: "Test", Priority: "medium"})

	found, err := tr.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestTaskFindAll(t *testing.T) {
	tr := setupTaskTestDB(t)
	ctx := context.Background()

	tr.Create(ctx, model.CreateTaskInput{Title: "Task 1", Priority: "low"})
	tr.Create(ctx, model.CreateTaskInput{Title: "Task 2", Priority: "high"})

	tasks, err := tr.FindAll(ctx, model.TaskFilters{})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestTaskFindAll_ExcludesCompleted(t *testing.T) {
	tr := setupTaskTestDB(t)
	ctx := context.Background()

	t1, _ := tr.Create(ctx, model.CreateTaskInput{Title: "Task 1", Priority: "low"})
	tr.Create(ctx, model.CreateTaskInput{Title: "Task 2", Priority: "high"})
	tr.Complete(ctx, t1.ID)

	tasks, err := tr.FindAll(ctx, model.TaskFilters{})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	// Include completed
	tasks, err = tr.FindAll(ctx, model.TaskFilters{IncludeCompleted: true})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestTaskComplete(t *testing.T) {
	tr := setupTaskTestDB(t)
	ctx := context.Background()

	task, _ := tr.Create(ctx, model.CreateTaskInput{Title: "Test", Priority: "medium"})

	completed, err := tr.Complete(ctx, task.ID)
	require.NoError(t, err)
	assert.True(t, completed.Completed)
	assert.NotNil(t, completed.CompletedAt)
}

func TestTaskUpdate(t *testing.T) {
	tr := setupTaskTestDB(t)
	ctx := context.Background()

	task, _ := tr.Create(ctx, model.CreateTaskInput{Title: "Test", Priority: "medium"})

	newTitle := "Updated"
	newPriority := "high"
	updated, err := tr.Update(ctx, task.ID, model.UpdateTaskInput{
		Title:    &newTitle,
		Priority: &newPriority,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
	assert.Equal(t, "high", updated.Priority)
}

func TestTaskArchive(t *testing.T) {
	tr := setupTaskTestDB(t)
	ctx := context.Background()

	task, _ := tr.Create(ctx, model.CreateTaskInput{Title: "Test", Priority: "medium"})

	err := tr.Archive(ctx, task.ID)
	require.NoError(t, err)

	_, err = tr.FindByID(ctx, task.ID)
	assert.ErrorIs(t, err, model.ErrNotFound)
}
