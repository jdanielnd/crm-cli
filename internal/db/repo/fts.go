package repo

import "strings"

// escapeFTS quotes a user query for safe use in FTS5 MATCH expressions.
// It wraps the query in double quotes (escaping any internal quotes)
// and appends a prefix wildcard (*) for partial matching.
func escapeFTS(query string) string {
	return `"` + strings.ReplaceAll(query, `"`, `""`) + `"*`
}

// defaultLimit returns the given limit, or 20 if it is zero or negative.
func defaultLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	return limit
}
