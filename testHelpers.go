package dxf

import (
	"strings"
	"testing"
)

func assertContains(t *testing.T, expected, actual string) {
	if !strings.Contains(actual, expected) {
		t.Errorf("Unable to find '%s' in '%s'", expected, actual)
	}
}

func assertEq(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}
