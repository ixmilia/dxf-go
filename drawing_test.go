package dxf

import "testing"

func TestDefaultDrawingVersion(t *testing.T) {
	drawing := *NewDrawing()
	actual := drawing.String()
	expected := "  1\r\nAC1009\r\n"
	assertContains(t, expected, actual)
}
