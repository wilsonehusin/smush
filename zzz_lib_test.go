package smush_test

import "testing"

func compareValues[T comparable](t *testing.T, expected, received T, reference any) {
	if expected != received {
		t.Fatalf("expected '%v', received '%v':\n%+v", expected, received, reference)
	}
}
