package dxf

// ControlPoint represents a control point of a Spline
type ControlPoint struct {
	Point  Point
	Weight float64
}

// NewControlPoint creates a new ControlPoint for a Spline
func NewControlPoint() *ControlPoint {
	return &ControlPoint{
		Point:  *NewOrigin(),
		Weight: 1.0,
	}
}
