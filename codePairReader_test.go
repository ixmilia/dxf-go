package dxf

import (
	"strings"
	"testing"
)

func TestReadEmptyFile(t *testing.T) {
	_ = parse(t, "")
}

func TestReadFileNewlines(t *testing.T) {
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
