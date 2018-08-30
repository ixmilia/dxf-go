package dxf

import "testing"

func assertEq(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}
