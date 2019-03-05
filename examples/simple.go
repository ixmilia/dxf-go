package examples

import (
	"fmt"

	dxf "github.com/IxMilia/dxf-go"
)

func read() {
	drawing, err := dxf.ReadFile("path/to/file.dxf")
	if err != nil {
		fmt.Printf("Error reading file: %s", err)
	}

	for _, entity := range drawing.Entities {
		switch ent := entity.(type) {
		case *dxf.Line:
			fmt.Printf("Line from %s to %s\n", ent.P1.String(), ent.P2.String())
		}
	}
}

func write() {
	drawing := *dxf.NewDrawing()
	line := dxf.NewLine()
	line.P1 = dxf.Point{X: 1.0, Y: 2.0, Z: 3.0}
	line.P2 = dxf.Point{X: 4.0, Y: 5.0, Z: 6.0}
	drawing.Entities = append(drawing.Entities, line)

	err := drawing.SaveFile("path/to/file.dxf")
	if err != nil {
		fmt.Printf("Error saving file: %s", err)
	}
}
