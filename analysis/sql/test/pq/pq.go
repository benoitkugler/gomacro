// Package pq is used as replacement for "github.com/lib/pq" in tests
package pq

import "database/sql/driver"

type Int64Array []int64

func CopyIn(name string, args ...string) string {
	return ""
}

type NullInt64 struct {
	Valid bool
	Int64 int64
}

func (s *NullInt64) Scan(src interface{}) error  { return nil }
func (s NullInt64) Value() (driver.Value, error) { return nil, nil }
