package sql

import (
	"regexp"
	"strings"
)

var (
	// check if the constraint is of the form UNIQUE(col1, col2, ...)
	reUniqueOrPrimary = regexp.MustCompile(`(?i)ADD (UNIQUE|PRIMARY KEY)\s?\((.*)\)`)
	reSelectKey       = regexp.MustCompile(`(?i)_SELECT KEY\s?\((.*)\)`)
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
		cols := strings.Split(matchs[2], ",")
		for i, c := range cols {
			cols[i] = strings.TrimSpace(c)
		}
		return cols
	}
	return nil
}

// matches _SELECT KEY (col1, col2, ...)
func isSelectKey(ct string) []string {
	matchs := reSelectKey.FindStringSubmatch(ct)
	if len(matchs) > 0 {
		cols := strings.Split(matchs[1], ",")
		for i, c := range cols {
			cols[i] = strings.TrimSpace(c)
		}
		return cols
	}
	return nil
}
