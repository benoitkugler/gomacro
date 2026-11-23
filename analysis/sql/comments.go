package sql

import (
	"regexp"
	"strings"
)

var (
	// check if the constraint is of the form UNIQUE(col1, col2, ...)
	reUniqueOrPrimary = regexp.MustCompile(`(?i)ADD (UNIQUE|PRIMARY KEY)\s?\((.*)\)`)
	// check if the constraint is of the form
	// CREATE UNIQUE INDEX [name] ON table(col1, col2, ...) [WHERE ...]
	reUniqueIndex = regexp.MustCompile(`(?i)CREATE\s+UNIQUE\s+INDEX(?:\s+\S+)?\s+ON\s+[\w.]+\s*\(([^)]*)\)`)
	reSelectKey   = regexp.MustCompile(`(?i)_SELECT KEY\s?\((.*)\)`)
)

// return the column made unique by a UNIQUE(col) constraint
// or the empty string
func isUniqueConstraint(ct string) string {
	if cols := isUniquesConstraint(ct); len(cols) == 1 {
		return cols[0]
	}
	return ""
}

// matches a UNIQUE(col1, col2, ...) or PRIMARY KEY(col1, col2, ...) constraint,
// returning the columns, or an empty slice.
func isUniquesConstraint(ct string) []string {
	matchs := reUniqueOrPrimary.FindStringSubmatch(ct)
	if len(matchs) > 0 {
		return splitColumns(matchs[2])
	}
	return nil
}

// matches a CREATE UNIQUE INDEX [name] ON table(col1, col2, ...) constraint,
// ignoring any WHERE clause, and returns the indexed columns.
func isUniqueIndexColumns(ct string) []string {
	matchs := reUniqueIndex.FindStringSubmatch(ct)
	if len(matchs) > 0 {
		return splitColumns(matchs[1])
	}
	return nil
}

// matches _SELECT KEY (col1, col2, ...)
func isSelectKey(ct string) []string {
	matchs := reSelectKey.FindStringSubmatch(ct)
	if len(matchs) > 0 {
		return splitColumns(matchs[1])
	}
	return nil
}

func splitColumns(cols string) []string {
	out := strings.Split(cols, ",")
	for i, c := range out {
		out[i] = strings.TrimSpace(c)
	}
	return out
}
