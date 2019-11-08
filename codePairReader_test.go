package dxf

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestReadEmptyFile(t *testing.T) {
	_ = parse(t, "")
}

func TestReadFileNewlines(t *testing.T) {
	vals := []string{
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$PROJECTNAME",
		"  1", "test",
		"  0", "ENDSEC",
		"  0", "EOF",
	}

	// parse with "\r\n" (standard)
	drawing := parse(t, strings.Join(vals, "\r\n"))
	assertEqString(t, "test", drawing.Header.ProjectName)

	// parse with "\n" (non-standard, but still acceptable)
	drawing = parse(t, strings.Join(vals, "\n"))
	assertEqString(t, "test", drawing.Header.ProjectName)
}

func TestReadTextAsAscii(t *testing.T) {
	// if version <= R2004 (AC1018) stream is ASCII

	// unicode values in the middle of the string
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$ACADVER",
		"  1", "AC1018",
		"  9", "$PROJECTNAME",
		"  1", "Rep\\U+00E8re pi\\U+00E8ce",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqString(t, "Repère pièce", drawing.Header.ProjectName)

	// unicode values for the entire string
	drawing = parse(t, join(
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$ACADVER",
		"  1", "AC1018",
		"  9", "$PROJECTNAME",
		"  1", "\\U+4F60\\U+597D",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqString(t, "你好", drawing.Header.ProjectName)
}

func TestReadTextAsUtf8(t *testing.T) {
	// if version >= R2007 (AC1021) stream is UTF-8
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$ACADVER",
		"  1", "AC1021",
		"  9", "$PROJECTNAME",
		"  1", "Repère pièce",
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqString(t, "Repère pièce", drawing.Header.ProjectName)
}

func TestReadBinaryFile(t *testing.T) {
	// pre R13 binary file
	drawing, err := ReadFile("res/diamond-bin.dxf")
	if err != nil {
		t.Error(err)
	}
	assertEqInt(t, 12, len(drawing.Entities))
	switch line := drawing.Entities[0].(type) {
	case *Line:
		assertEqPoint(t, Point{45.0, 45.0, 0.0}, line.P1)
		assertEqPoint(t, Point{45.0, -45.0, 0.0}, line.P2)
	default:
		t.Error("expected LINE")
	}
}

func TestReadBinaryPostR13(t *testing.T) {
	data := make([]byte, 0)
	// binary header
	data = append(data, []byte("AutoCAD Binary DXF\r\n")...)
	data = append(data, []byte{0x1A, 0x00}...)

	// 0/SECTION
	data = append(data, []byte{0x00, 0x00}...)
	data = append(data, []byte("SECTION")...)
	data = append(data, 0x00)

	// 2/HEADER
	data = append(data, []byte{0x02, 0x00}...)
	data = append(data, []byte("HEADER")...)
	data = append(data, 0x00)

	// 9/$LWDISPLAY
	data = append(data, []byte{0x09, 0x00}...)
	data = append(data, []byte("$LWDISPLAY")...)
	data = append(data, 0x00)

	// 290/true
	data = append(data, []byte{0x22, 0x01}...)
	data = append(data, 0x01)

	// 0/ENDSEC
	data = append(data, []byte{0x00, 0x00}...)
	data = append(data, []byte("ENDSEC")...)
	data = append(data, 0x00)

	// 0/EOF
	data = append(data, []byte{0x00, 0x00}...)
	data = append(data, []byte("EOF")...)
	data = append(data, 0x00)

	reader := bytes.NewReader(data)
	drawing, err := ReadFromReader(reader)
	if err != nil {
		t.Error(err)
	}

	assert(t, drawing.Header.DisplayLinewieghtInModelAndLayoutTab, "expected $LWDISPLAY to be true")
}

func TestReadDrawingWithNonStandardEncoding(t *testing.T) {
	contents := join(
		"  0", "SECTION",
		"  2", "HEADER",
		"  9", "$PROJECTNAME",
		"  1", "\xB2\xBB",
		"  0", "ENDSEC",
		"  0", "EOF",
	)
	data := []byte(contents)
	reader := bytes.NewReader(data)
	drawing, err := ReadFromReaderWithEncoding(reader, simplifiedchinese.GB18030)
	if err != nil {
		t.Error(err)
	}

	assertEqString(t, "不", drawing.Header.ProjectName)
}
