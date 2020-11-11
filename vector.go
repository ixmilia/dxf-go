package dxf

import (
	"fmt"
)

// The Vector struct represents a vector in 3D space.
type Vector struct {
	X float64
	Y float64
	Z float64
}

// NewXAxis creates a unit vector along the X axis.
func NewXAxis() *Vector {
	return &Vector{
		X: 1.0,
		Y: 0.0,
		Z: 0.0,
	}
}

// NewYAxis creates a unit vector along the Y axis.
func NewYAxis() *Vector {
	return &Vector{
		X: 0.0,
		Y: 1.0,
		Z: 0.0,
	}
}

// NewZAxis creates a unit vector along the Z axis.
func NewZAxis() *Vector {
	return &Vector{
		X: 0.0,
		Y: 0.0,
		Z: 1.0,
	}
}

// NewZeroVector creates a vector representing zero distance across any axis.
func NewZeroVector() *Vector {
	return &Vector{
		X: 0.0,
		Y: 0.0,
		Z: 0.0,
	}
}

func (v *Vector) String() string {
	return fmt.Sprintf("(%s, %s, %s)", formatFloat64Text(v.X), formatFloat64Text(v.Y), formatFloat64Text(v.Z))
}
