package dxf

import (
	"strings"
	"testing"
)

func assert(t *testing.T, condition bool, message string) {
	if !condition {
		t.Error(message)
	}
}

func assertContains(t *testing.T, expected, actual string) {
	if !strings.Contains(actual, expected) {
		t.Errorf("Unable to find '%s' in '%s'", expected, actual)
	}
}

func assertNotContains(t *testing.T, notExpected, actual string) {
	if strings.Contains(actual, notExpected) {
		t.Errorf("Unexpectedly found '%s' in '%s'", notExpected, actual)
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

func join(vals ...string) string {
	return strings.Join(vals, "\r\n")
}

func parse(t *testing.T, content string) Drawing {
	drawing, err := ParseDrawing(strings.TrimSpace(content))
	if err != nil {
		t.Error(err)
	}

	return drawing
}
