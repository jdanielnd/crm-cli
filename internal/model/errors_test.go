package model_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jdanielnd/crm-cli/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"not found", model.ErrNotFound, 3},
		{"validation", model.ErrValidation, 2},
		{"conflict", model.ErrConflict, 4},
		{"database", model.ErrDatabase, 10},
		{"generic", errors.New("something"), 1},
		{"wrapped not found", fmt.Errorf("person 42: %w", model.ErrNotFound), 3},
		{"exit error wrapping not found", model.NewExitError(model.ErrNotFound, "person %d not found", 42), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, model.ExitCode(tt.err))
		})
	}
}

func TestExitError(t *testing.T) {
	err := model.NewExitError(model.ErrNotFound, "person %d not found", 42)
	assert.Equal(t, "person 42 not found", err.Error())
	assert.True(t, errors.Is(err, model.ErrNotFound))
	assert.False(t, errors.Is(err, model.ErrValidation))
}
