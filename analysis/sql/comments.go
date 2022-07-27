package sql

import (
	"regexp"
	"strings"
)

// check if the constraint is of the form UNIQUE(col1, col2, ...)
var reUnique = regexp.MustCompile(`(?i)ADD UNIQUE\s?\((.*)\)`)

// return the column made unique by a UNIQUE(col) constraint
// or the empty string
func isUniqueConstraint(ct string) string {
	matchs := reUnique.FindStringSubmatch(ct)
	if len(matchs) > 0 {
		cols := strings.Split(matchs[1], ",")
		if len(cols) == 1 { // unique column
			return strings.TrimSpace(cols[0])
		}
	}
	return ""
}
