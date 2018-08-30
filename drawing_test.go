package dxf

import "testing"

func TestEmptyDrawing(t *testing.T) {
	drawing := *NewDrawing()
	actual := drawing.String()
	expected := "  0\r\nEOF\r\n"
	assertEq(t, expected, actual)
}
