package dxf

// LwVertex represents a vertex of an LWPolyline
type LwVertex struct {
	X             float64
	Y             float64
	ID            int
	StartingWidth float64
	EndingWidth   float64
	Bulge         float64
}

// NewLwVertex creates a new LwVertex for an LWPolyline
func NewLwVertex() *LwVertex {
	return &LwVertex{
		X:             0.0,
		Y:             0.0,
		ID:            0,
		StartingWidth: 0.0,
		EndingWidth:   0.0,
		Bulge:         0.0,
	}
}
