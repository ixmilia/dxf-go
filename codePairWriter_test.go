package dxf

import (
	"bufio"
	"bytes"
	"testing"
)

func TestWriteNewLines(t *testing.T) {
	str := NewDrawing().String()
	assertNotContains(t, "9\n$ACADVER", str)
	assertContains(t, "9\r\n$ACADVER", str)
}

func TestWriteTextAsAscii(t *testing.T) {
	// if version <= R2004 stream is ASCII
	drawing := *NewDrawing()
	drawing.Header.Version = R2004
	drawing.Header.ProjectName = "Repère pièce"

	actual := drawing.String()
	expected := join(
		"  9", "$PROJECTNAME",
		"  1", "Rep\\U+00E8re pi\\U+00E8ce")
	assertContains(t, expected, actual)
}

func TestWriteTextAsUtf8(t *testing.T) {
	// if version >= R2007 (AC1018) stream is UTF-8
	drawing := *NewDrawing()
	drawing.Header.Version = R2007
	drawing.Header.ProjectName = "Repère pièce"

	actual := drawing.String()
	expected := join(
		"  9", "$PROJECTNAME",
		"  1", "Repère pièce")
	assertContains(t, expected, actual)
}

func TestRoundTripBinaryFile(t *testing.T) {
	for _, version := range []AcadVersion{R12, R13} {
		t.Logf("Testing binary roundtrip version %s", version.String())
		drawing := *NewDrawing()
		drawing.Header.Version = version
		drawing.Header.CurrentLayer = "current-layer"
		buf := new(bytes.Buffer)
		err := drawing.SaveToWriterBinary(buf)
		if err != nil {
			t.Error(err)
		}
		bs := buf.Bytes()
		reader := bytes.NewReader(bs)
		drawing, err = ReadFromReader(reader)
		if err != nil {
			t.Error(err)
		}
		assertEqString(t, "current-layer", drawing.Header.CurrentLayer)
	}
}

func TestWriteBinary(t *testing.T) {
	for _, version := range []AcadVersion{R12, R13} {
		t.Logf("Testing binary write version %s", version.String())
		drawing := *NewDrawing()
		drawing.Header.Version = version
		buf := new(bytes.Buffer)
		writer := bufio.NewWriter(buf)
		err := drawing.SaveToWriterBinary(writer)
		if err != nil {
			t.Error(err)
		}
		err = writer.Flush()
		if err != nil {
			t.Error(err)
		}
		bts := buf.Bytes()
		binarySentinelBits := bts[0:20]
		binarySentinel := string(binarySentinelBits)
		assertEqString(t, "AutoCAD Binary DXF\r\n", binarySentinel)
		var expectedSectionText []byte
		if version < R13 {
			expectedSectionText = bts[23:30]
		} else {
			expectedSectionText = bts[24:31]
		}
		actual := string(expectedSectionText)
		assertEqString(t, "SECTION", actual)
	}
}
