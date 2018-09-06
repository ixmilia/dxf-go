package dxf

import (
	"fmt"
	"testing"
)

func TestDefaultDrawingVersion(t *testing.T) {
	drawing := *NewDrawing()
	actual := drawing.String()
	expected := join(
		"  1", "AC1009")
	assertContains(t, expected, actual)
}

func TestReadingUnsupportedSections(t *testing.T) {
	drawing := parse(t, join(
		// unsupported
		"  0", "SECTION",
		"  2", "NOT_ENTITIES",
		"  0", "NOT_AN_ENTITY",
		"  0", "ENDSEC",
		// header
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$ACADVER",
		"  1", "AC1015",
		"  0", "ENDSEC",
		// unsupported
		"  0", "SECTION",
		"  2", "NOT_ENTITIES_AGAIN",
		"  0", "STILL_NOT_AN_ENTITY",
		"  0", "ENDSEC",
		// end
		"  0", "EOF",
	))
	expected := R2000
	assert(t, drawing.Header.Version == R2000, fmt.Sprintf("Expected: %v\nActual: %v", expected, drawing.Header.Version))
}
