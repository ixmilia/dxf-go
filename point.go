package dxf

import (
	"fmt"
)

// The Point struct represents a 3D coordinate.
type Point struct {
	X float64
	Y float64
	Z float64
}

// NewOrigin creates a new Point representing the (0, 0, 0) location.
func NewOrigin() *Point {
	return &Point{
		X: 0.0,
		Y: 0.0,
		Z: 0.0,
	}
}

func (p *Point) String() string {
	return fmt.Sprintf("(%s, %s, %s)", formatFloat64Text(p.X), formatFloat64Text(p.Y), formatFloat64Text(p.Z))
}
