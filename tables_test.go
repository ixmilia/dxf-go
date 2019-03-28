package dxf

import (
	"strings"
	"testing"
)

func TestReadLayer(t *testing.T) {
	drawing := parseTableItem(t, "LAYER", join(
		"  2", "layer-name",
	))
	assertEqInt(t, 1, len(drawing.Layers))
	layer := drawing.Layers[0]
	assertEqString(t, "layer-name", layer.Name)
}

func TestWriteLayer(t *testing.T) {
	l := *NewLayer()
	l.Name = "layer-name"
	d := NewDrawing()
	d.Layers = append(d.Layers, l)
	actual := d.String()
	assertContains(t, join(
		"100", "AcDbSymbolTableRecord",
		"100", "AcDbLayerTableRecord",
		"  2", "layer-name",
		" 62", "     7",
		"  6", "CONTINUOUS",
	), actual)
}

func TestRoundTripLayer(t *testing.T) {
	l := *NewLayer()
	l.Name = "layer-name"
	d := NewDrawing()
	d.Layers = append(d.Layers, l)
	r := roundTripDrawing(t, d)
	assertEqInt(t, 1, len(r.Layers))
	l2 := r.Layers[0]
	assertEqString(t, l.Name, l2.Name)
}

func TestReadLayers(t *testing.T) {
	drawing := parse(t, join(
		// section decl
		"  0", "SECTION",
		"  2", "TABLES",
		// table decl
		"  0", "TABLE",
		"  2", "LAYER",
		// item
		"  0", "LAYER",
		"  2", "layer-1",
		// item
		"  0", "LAYER",
		"  2", "layer-2",
		// end
		"  0", "ENDTAB",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqInt(t, 2, len(drawing.Layers))
	assertEqString(t, "layer-1", drawing.Layers[0].Name)
	assertEqString(t, "layer-2", drawing.Layers[1].Name)
}

func TestUnsupportedTable(t *testing.T) {
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "TABLES",
		"  0", "TABLE",
		"  2", "UNSUPPORTED",
		"  0", "UNSUPPORTED",
		"  2", "unsupported-name",
		"  0", "ENDTAB",
		"  0", "TABLE",
		"  2", "LAYER",
		"  0", "LAYER",
		"  2", "layer-name",
		"  0", "ENDTAB",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqInt(t, 1, len(drawing.Layers))
	assertEqString(t, "layer-name", drawing.Layers[0].Name)
}

func parseTableItem(t *testing.T, tableType, body string) (drawing Drawing) {
	drawing = parse(t, join(
		"  0", "SECTION",
		"  2", "TABLES",
		"  0", "TABLE",
		"  2", tableType,
		"  0", tableType,
	)+"\r\n"+strings.TrimSpace(body)+"\r\n"+join(
		"  0", "ENDTAB",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	return
}
