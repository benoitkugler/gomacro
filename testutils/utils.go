package testutils

import "testing"

func ShouldPanic(t *testing.T, f func()) {
	t.Helper()

	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}

func Assert(t *testing.T, b bool, context ...interface{}) {
	t.Helper()
	if !b {
		if len(context) >= 1 {
			t.Fatalf("assertion error %v", context[0])
		} else {
			t.Fatalf("assertion error")
		}
	}
}
