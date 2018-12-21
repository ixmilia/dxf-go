package dxf

type Vertex struct {
	X             float64
	Y             float64
	ID            int
	StartingWidth float64
	EndingWidth   float64
	Bulge         float64
}

func NewVertex() *Vertex {
	return &Vertex{
		X:             0.0,
		Y:             0.0,
		ID:            0,
		StartingWidth: 0.0,
		EndingWidth:   0.0,
		Bulge:         0.0,
	}
}
