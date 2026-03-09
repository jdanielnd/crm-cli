package cli

import (
	"strconv"

	"github.com/jdanielnd/crm-cli/internal/model"
)

// parseEntityID parses a string as an int64 entity ID and returns
// a validation error with the entity name on failure.
func parseEntityID(s string, entity string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, model.NewExitError(model.ErrValidation, "invalid %s ID: %s", entity, s)
	}
	return id, nil
}

// parseEntityIDs parses multiple string arguments as int64 entity IDs.
func parseEntityIDs(args []string, entity string) ([]int64, error) {
	ids := make([]int64, 0, len(args))
	for _, arg := range args {
		id, err := parseEntityID(arg, entity)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
