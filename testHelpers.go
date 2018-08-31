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

func assertEqInt(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected: %d\nActual: %d", expected, actual)
	}
}

func assertEqString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}

func parse(t *testing.T, content string) Drawing {
	drawing, err := ParseDrawing(strings.TrimSpace(content))
	if err != nil {
		t.Error(err)
	}

	return drawing
}
