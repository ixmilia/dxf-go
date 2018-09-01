package dxf

import (
	"fmt"
)

type Vector struct {
	X float64
	Y float64
	Z float64
}

func NewXAxis() *Vector {
	return &Vector{
		X: 1.0,
		Y: 0.0,
		Z: 0.0,
	}
}

func NewYAxis() *Vector {
	return &Vector{
		X: 0.0,
		Y: 1.0,
		Z: 0.0,
	}
}

func NewZAxis() *Vector {
	return &Vector{
		X: 0.0,
		Y: 0.0,
		Z: 1.0,
	}
}

func NewZeroVector() *Vector {
	return &Vector{
		X: 0.0,
		Y: 0.0,
		Z: 0.0,
	}
}

func (v *Vector) String() string {
	return fmt.Sprintf("(%f, %f, %f)", v.X, v.Y, v.Z)
}
