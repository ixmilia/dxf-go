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

func TestRoundTripAllDrawingVersions(t *testing.T) {
	// testing with the useful subset of header versions
	versions := []AcadVersion{
		R9,
		R10,
		R12,
		R13,
		R14,
		R2000,
		R2004,
		R2007,
		R2010,
		R2013,
		R2018,
	}
	for _, version := range versions {
		t.Logf("Testing version roundtrip %s", version.String())
		drawing := *NewDrawing()
		drawing.Header.Version = version
		line := NewLine()
		p1 := Point{1.0, 2.0, 3.0}
		p2 := Point{4.0, 5.0, 6.0}
		line.P1 = p1
		line.P2 = p2
		drawing.Entities = append(drawing.Entities, line)

		parsedDrawing := parse(t, drawing.String())
		assert(t, parsedDrawing.Header.Version == version, fmt.Sprintf("Expected: %s\nActual: %s", parsedDrawing.Header.Version.String(), version.String()))
		assertEqInt(t, 1, len(parsedDrawing.Entities))
		assertEqString(t, "LINE", parsedDrawing.Entities[0].typeString())
		parsedLine, _ := parsedDrawing.Entities[0].(*Line)
		assert(t, parsedLine.P1 == p1, fmt.Sprintf("Expected: %s\nActual: %s", p1.String(), parsedLine.P1.String()))
		assert(t, parsedLine.P2 == p2, fmt.Sprintf("Expected: %s\nActual: %s", p2.String(), parsedLine.P2.String()))
	}
}

func TestReadTextAsAscii(t *testing.T) {
	// if version <= R2004 (AC1018) stream is ASCII

	// unicode values in the middle of the string
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$ACADVER",
		"  1", "AC1018",
		"  9", "$PROJECTNAME",
		"  1", "Rep\\U+00E8re pi\\U+00E8ce",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqString(t, "Repère pièce", drawing.Header.ProjectName)

	// unicode values for the entire string
	drawing = parse(t, join(
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$ACADVER",
		"  1", "AC1018",
		"  9", "$PROJECTNAME",
		"  1", "\\U+4F60\\U+597D",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqString(t, "你好", drawing.Header.ProjectName)
}

func TestReadTextAsUtf8(t *testing.T) {
	// if version >= R2007 (AC1021) stream is UTF-8
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$ACADVER",
		"  1", "AC1021",
		"  9", "$PROJECTNAME",
		"  1", "Repère pièce",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqString(t, "Repère pièce", drawing.Header.ProjectName)
}

func TestWriteTextAsAscii(t *testing.T) {
	// if version <= R2004 stream is ASCII
	drawing := *NewDrawing()
	drawing.Header.Version = R2004
	drawing.Header.ProjectName = "Repère pièce"

	actual := drawing.String()
	expected := join(
		"  9", "$PROJECTNAME",
		"  1", "Rep\\U+00E8re pi\\U+00E8ce")
	assertContains(t, expected, actual)
}

func TestWriteTextAsUtf8(t *testing.T) {
	// if version >= R2007 (AC1018) stream is UTF-8
	drawing := *NewDrawing()
	drawing.Header.Version = R2007
	drawing.Header.ProjectName = "Repère pièce"

	actual := drawing.String()
	expected := join(
		"  9", "$PROJECTNAME",
		"  1", "Repère pièce")
	assertContains(t, expected, actual)
}
