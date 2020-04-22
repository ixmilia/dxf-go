package dxf

import "testing"

func TestSimpleLineCanBeReadByOda(t *testing.T) {
	oda, err := newOdaHelper()
	if err != nil {
		return
	}

	defer oda.cleanup(t)

	drawing := *NewDrawing()
	line := NewLine()
	line.P1 = Point{X: 1.0, Y: 2.0, Z: 3.0}
	line.P2 = Point{X: 4.0, Y: 5.0, Z: 6.0}
	drawing.Entities = append(drawing.Entities, line)

	roundTripped := oda.convertDrawing(t, &drawing, R2000)
	assertEqInt(t, 1, len(roundTripped.Entities))
	roundTrippedLine := roundTripped.Entities[0].(*Line)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, roundTrippedLine.P1)
	assertEqPoint(t, Point{4.0, 5.0, 6.0}, roundTrippedLine.P2)
}
