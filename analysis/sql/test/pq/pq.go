// Package pq is used as replacement for "github.com/lib/pq" in tests
package pq

import "database/sql/driver"

type Int64Array []int64
type Int32Array []int32

func (*Int64Array) Scan(src interface{}) error  { return nil }
func (*Int32Array) Scan(src interface{}) error  { return nil }
func (Int64Array) Value() (driver.Value, error) { return nil, nil }
func (Int32Array) Value() (driver.Value, error) { return nil, nil }

func CopyIn(name string, args ...string) string {
	return ""
}
