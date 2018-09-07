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
