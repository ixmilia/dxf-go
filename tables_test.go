package dxf

import (
	"testing"
)

func TestReadLayer(t *testing.T) {
	drawing := parseTableItem(t, "LAYER",
		NewStringCodePair(2, "layer-name"),
	)
	assertEqInt(t, 1, len(drawing.Layers))
	layer := drawing.Layers[0]
	assertEqString(t, "layer-name", layer.Name)
}

func TestWriteLayer(t *testing.T) {
	l := *NewLayer()
	l.Name = "layer-name"
	d := NewDrawing()
	d.Layers = append(d.Layers, l)
	actual, err := d.CodePairs()
	if err != nil {
		t.Error(err)
	}
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDbSymbolTableRecord"),
		NewStringCodePair(2, "layer-name"),
		NewShortCodePair(70, 0),
		NewShortCodePair(62, 7),
		NewStringCodePair(6, "CONTINUOUS"),
	}, actual)
}

func TestRoundTripLayer(t *testing.T) {
	l := *NewLayer()
	l.Name = "layer-name"
	d := NewDrawing()
	d.Layers = append(d.Layers, l)
	r := roundTripDrawing(t, d)
	var l2 *Layer
	for i := range r.Layers {
		l2 = &r.Layers[i]
		if l2.Name == "layer-name" {
			break
		}
	}

	if l2 == nil {
		t.Errorf("Layer not found in round-tripped drawing")
	}
}

func TestReadLayers(t *testing.T) {
	drawing := parseFromCodePairs(t,
		// section decl
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "TABLES"),
		// table decl
		NewStringCodePair(0, "TABLE"),
		NewStringCodePair(2, "LAYER"),
		// item
		NewStringCodePair(0, "LAYER"),
		NewStringCodePair(2, "layer-1"),
		// item
		NewStringCodePair(0, "LAYER"),
		NewStringCodePair(2, "layer-2"),
		// end
		NewStringCodePair(0, "ENDTAB"),
		NewStringCodePair(0, "ENDSEC"),
		NewStringCodePair(0, "EOF"),
	)
	assertEqInt(t, 2, len(drawing.Layers))
	assertEqString(t, "layer-1", drawing.Layers[0].Name)
	assertEqString(t, "layer-2", drawing.Layers[1].Name)
}

func TestReadTableWithHandle(t *testing.T) {
	drawing := parseFromCodePairs(t,
		// section decl
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "TABLES"),
		// table decl
		NewStringCodePair(0, "TABLE"),
		NewStringCodePair(2, "VPORT"),
		NewStringCodePair(5, "ABCD"), // n.b., handle is on the table, not the table item
		// item
		NewStringCodePair(0, "VPORT"),
		NewStringCodePair(2, "vport-name"),
		// end
		NewStringCodePair(0, "ENDTAB"),
		NewStringCodePair(0, "ENDSEC"),
		NewStringCodePair(0, "EOF"),
	)
	assertEqInt(t, 1, len(drawing.ViewPorts))
	assertEqString(t, "vport-name", drawing.ViewPorts[0].Name)
}

func TestUnsupportedTable(t *testing.T) {
	drawing := parseFromCodePairs(t,
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "TABLES"),
		NewStringCodePair(0, "TABLE"),
		NewStringCodePair(2, "UNSUPPORTED"),
		NewStringCodePair(0, "UNSUPPORTED"),
		NewStringCodePair(2, "unsupported-name"),
		NewStringCodePair(0, "ENDTAB"),
		NewStringCodePair(0, "TABLE"),
		NewStringCodePair(2, "LAYER"),
		NewStringCodePair(0, "LAYER"),
		NewStringCodePair(2, "layer-name"),
		NewStringCodePair(0, "ENDTAB"),
		NewStringCodePair(0, "ENDSEC"),
		NewStringCodePair(0, "EOF"),
	)
	assertEqInt(t, 1, len(drawing.Layers))
	assertEqString(t, "layer-name", drawing.Layers[0].Name)
}

func parseTableItem(t *testing.T, tableType string, codePairs ...CodePair) (drawing Drawing) {
	allPairs := []CodePair{
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "TABLES"),
		NewStringCodePair(0, "TABLE"),
		NewStringCodePair(2, tableType),
		NewStringCodePair(0, tableType),
	}
	allPairs = append(allPairs, codePairs...)
	allPairs = append(allPairs,
		NewStringCodePair(0, "ENDTAB"),
		NewStringCodePair(0, "ENDSEC"),
		NewStringCodePair(0, "EOF"),
	)
	drawing = parseFromCodePairs(t, allPairs...)
	return
}
