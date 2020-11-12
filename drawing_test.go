package dxf

import (
	"fmt"
	"strings"
	"testing"
)

func TestDefaultDrawingVersion(t *testing.T) {
	drawing := *NewDrawing()
	actual, err := drawing.CodePairs()
	if err != nil {
		t.Error(err)
	}
	expected := []CodePair{
		NewStringCodePair(1, "AC1009"),
	}
	assertContainsCodePairs(t, expected, actual)
}

func TestReadTextFileNewlines(t *testing.T) {
	vals := []string{
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$PROJECTNAME",
		"  1", "test",
		"  0", "ENDSEC",
		"  0", "EOF",
	}

	// parse with "\r\n" (standard)
	drawing := parse(t, strings.Join(vals, "\r\n"))
	assertEqString(t, "test", drawing.Header.ProjectName)

	// parse with "\n" (non-standard, but still acceptable)
	drawing = parse(t, strings.Join(vals, "\n"))
	assertEqString(t, "test", drawing.Header.ProjectName)
}

func TestReadingUnsupportedSections(t *testing.T) {
	drawing := parseFromCodePairs(t,
		// unsupported
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "NOT_ENTITIES"),
		NewStringCodePair(0, "NOT_AN_ENTITY"),
		NewStringCodePair(0, "ENDSEC"),
		// header
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "HEADER"),
		NewStringCodePair(9, "$ACADVER"),
		NewStringCodePair(1, "AC1015"),
		NewStringCodePair(0, "ENDSEC"),
		// unsupported
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "NOT_ENTITIES_AGAIN"),
		NewStringCodePair(0, "STILL_NOT_AN_ENTITY"),
		NewStringCodePair(0, "ENDSEC"),
		// end
		NewStringCodePair(0, "EOF"),
	)
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

func TestReadAndNavigateHandles(t *testing.T) {
	drawing := parseFromCodePairs(t,
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "ENTITIES"),
		// artificial parent circle
		NewStringCodePair(0, "CIRCLE"),
		NewStringCodePair(5, "9999"), // handle
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		// artificial child line
		NewStringCodePair(0, "LINE"),
		NewStringCodePair(330, "9999"), // ownerhandle
		NewStringCodePair(0, "ENDSEC"),
		NewStringCodePair(0, "EOF"),
	)
	line := drawing.Entities[len(drawing.Entities)-1].(*Line)
	circle := (*line.Owner()).(*Circle)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, circle.Center)
}

func TestRoundTripOwnerPointers(t *testing.T) {
	// line sets circle as its owner
	circle := NewCircle()
	circle.Center = Point{1.0, 2.0, 3.0}
	parent := DrawingItem(circle)
	line := NewLine()
	line.SetOwner(&parent)
	drawing := *NewDrawing()
	drawing.Header.Version = R13 // handles only written on >= R13
	drawing.Entities = append(drawing.Entities, line)
	drawing.Entities = append(drawing.Entities, circle)
	drawingString := drawing.String()

	// verify circle is still owner of line
	reParsedDrawing := parse(t, drawingString)
	reParsedLine := reParsedDrawing.Entities[0].(*Line)
	reParsedCircle := (*reParsedLine.Owner()).(*Circle)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, reParsedCircle.Center)
}
