package dxf

import (
	"testing"
)

func TestWriteNewLines(t *testing.T) {
	str := NewDrawing().String()
	assertNotContains(t, "9\n$ACADVER", str)
	assertContains(t, "9\r\n$ACADVER", str)
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
