package dxf

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseSimpleLine(t *testing.T) {
	line := parseEntity(t, "LINE", join(
		" 10", "1.0",
		" 20", "2.0",
		" 30", "3.0",
		" 11", "4.0",
		" 21", "5.0",
		" 31", "6.0",
	)).(*Line)
	expectedP1 := Point{1.0, 2.0, 3.0}
	expectedP2 := Point{4.0, 5.0, 6.0}
	assert(t, expectedP1 == line.P1, fmt.Sprintf("Expected: %s\nActual: %s", expectedP1.String(), line.P1.String()))
	assert(t, expectedP2 == line.P2, fmt.Sprintf("Expected: %s\nActual: %s", expectedP2.String(), line.P2.String()))
}

func TestParseAlternateTypeString(t *testing.T) {
	line := parseEntity(t, "3DLINE", join(
		" 10", "1.0",
		" 20", "2.0",
		" 30", "3.0",
		" 11", "4.0",
		" 21", "5.0",
		" 31", "6.0",
	)).(*Line)
	expectedP1 := Point{1.0, 2.0, 3.0}
	expectedP2 := Point{4.0, 5.0, 6.0}
	assert(t, expectedP1 == line.P1, fmt.Sprintf("Expected: %s\nActual: %s", expectedP1.String(), line.P1.String()))
	assert(t, expectedP2 == line.P2, fmt.Sprintf("Expected: %s\nActual: %s", expectedP2.String(), line.P2.String()))
}

func TestParseUnsupportedEntity(t *testing.T) {
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "ENTITIES",
		"  0", "LINE", // supported entity
		"  0", "NOT_AN_ENTITY", // unsupported entity
		"  0", "LINE", // supported entity
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqInt(t, 2, len(drawing.Entities))
	assertEqString(t, "LINE", drawing.Entities[0].typeString())
	assertEqString(t, "LINE", drawing.Entities[1].typeString())
}

func TestWriteSimpleLine(t *testing.T) {
	line := NewLine()
	line.P1 = Point{1.0, 2.0, 3.0}
	line.P2 = Point{4.0, 5.0, 6.0}
	actual := entityString(line, R12)
	assertContains(t, join(
		"  0", "LINE",
	), actual)
	assertContains(t, join(
		" 10", "1.0",
		" 20", "2.0",
		" 30", "3.0",
		" 11", "4.0",
		" 21", "5.0",
		" 31", "6.0",
	), actual)
}

func TestConditionalEntityFieldWriting(t *testing.T) {
	line := NewLine()
	line.SetIsInPaperSpace(false)
	actual := entityString(line, R14)
	assertNotContains(t, join(
		"  0", "LINE",
		"100", "AcDbEntity",
		" 67", // [NO-VALUE] this is only written when Version >= R12 and it's not the default (false)
	), actual)
}

func TestReadEntityFieldFlag(t *testing.T) {
	face := parseEntity(t, "3DFACE", join(
		" 70", "     5",
	)).(*Face)
	assert(t, face.FirstEdgeInvisible(), "expected first edge to be invisible")
	assert(t, !face.SecondEdgeInvisible(), "expected first edge to be visible")
	assert(t, face.ThirdEdgeInvisible(), "expected first edge to be invisible")
	assert(t, !face.FourthEdgeInvisible(), "expected first edge to be visible")
}

func TestWriteEntityFieldFlag(t *testing.T) {
	face := NewFace()
	face.SetFirstEdgeInvisible(true)
	face.SetSecondEdgeInvisible(false)
	face.SetThirdEdgeInvisible(true)
	face.SetFourthEdgeInvisible(false)
	actual := entityString(face, R12)
	assertContains(t, join(
		// these parts just ensure we're checking the correct entity
		"100", "AcDbFace",
		" 10", "0.0",
		" 20", "0.0",
		" 30", "0.0",
		" 11", "0.0",
		" 21", "0.0",
		" 31", "0.0",
		" 12", "0.0",
		" 22", "0.0",
		" 32", "0.0",
		" 13", "0.0",
		" 23", "0.0",
		" 33", "0.0",
		// this is the real check
		" 70", "     5",
	), actual)
}

func TestWriteVersionSpecificEntities(t *testing.T) {
	solid := NewSolid()
	drawing := *NewDrawing()
	drawing.Entities = append(drawing.Entities, solid)

	// ensure it's present when appropriate
	drawing.Header.Version = R13
	assertContains(t, join(
		"  0", "3DSOLID",
	), drawing.String())

	// and not otherwise
	drawing.Header.Version = R12
	assertNotContains(t, join(
		"  0", "3DSOLID",
	), drawing.String())
}

func TestReadMultipleBaseEntityData(t *testing.T) {
	line := parseEntity(t, "LINE", join(
		"310", "line 1",
		"310", "line 2",
	)).(*Line)
	assertEqInt(t, 2, len(line.PreviewImageData()))
	assertEqString(t, "line 1", line.PreviewImageData()[0])
	assertEqString(t, "line 2", line.PreviewImageData()[1])
}

func TestWriteMultipleBaseEntityData(t *testing.T) {
	line := NewLine()
	line.AddPreviewImageData("line 1")
	line.AddPreviewImageData("line 2")
	actual := entityString(line, R2000)
	assertContains(t, join(
		"310", "line 1",
		"310", "line 2",
		"100", "AcDbLine",
	), actual)
}

func TestReadMultipleSpecificEntityData(t *testing.T) {
	solid := parseEntity(t, "3DSOLID", join(
		"  1", "line 1",
		"  1", "line 2",
	)).(*Solid)
	assertEqInt(t, 2, len(solid.CustomData))
	assertEqString(t, "line 1", solid.CustomData[0])
	assertEqString(t, "line 2", solid.CustomData[1])
}

func TestWriteMultipleSpecificEntityData(t *testing.T) {
	solid := NewSolid()
	solid.AddCustomData("line 1")
	solid.AddCustomData("line 2")
	actual := entityString(solid, R13)
	assertContains(t, join(
		"100", "AcDbModelerGeometry",
		" 70", "     1",
		"  1", "line 1",
		"  1", "line 2",
	), actual)
}

func parseEntity(t *testing.T, entityType string, body string) Entity {
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "ENTITIES",
		"  0", entityType,
	)+"\r\n"+strings.TrimSpace(body)+"\r\n"+join(
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqInt(t, 1, len(drawing.Entities))
	return drawing.Entities[0]
}

func entityString(entity Entity, version AcadVersion) string {
	drawing := *NewDrawing()
	drawing.Header.Version = version
	drawing.Entities = append(drawing.Entities, entity)
	return drawing.String()
}
