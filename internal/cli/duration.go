package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/jdanielnd/crm-cli/internal/model"
)

var durationRe = regexp.MustCompile(`^(\d+)([dwm])$`)

// parseDueDate parses a duration string like "7d", "2w", "1m" and returns
// an ISO 8601 date string relative to now.
// d = days, w = weeks, m = months.
func parseDueDate(s string) (string, error) {
	matches := durationRe.FindStringSubmatch(s)
	if matches == nil {
		return "", model.NewExitError(model.ErrValidation, "invalid duration %q (use format like 7d, 2w, 1m)", s)
	}

	n, _ := strconv.Atoi(matches[1])
	unit := matches[2]
	now := time.Now()

	var due time.Time
	switch unit {
	case "d":
		due = now.AddDate(0, 0, n)
	case "w":
		due = now.AddDate(0, 0, n*7)
	case "m":
		due = now.AddDate(0, n, 0)
	}

	return due.Format("2006-01-02"), nil
}

// derefString returns the string value or a default if the pointer is nil.
func derefString(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// fmtFollowupTitle creates a follow-up task title from an interaction.
func fmtFollowupTitle(subject *string, interactionType string) string {
	s := derefString(subject, interactionType)
	return fmt.Sprintf("Follow up: %s", s)
}
