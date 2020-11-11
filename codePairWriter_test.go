package dxf

import (
	"bufio"
	"bytes"
	"testing"
)

func TestWriteBoolAsText(t *testing.T) {
	assertText(t, "     1", formatBoolText(true))
	assertText(t, "     0", formatBoolText(false))
}

func TestWriteShortAsText(t *testing.T) {
	assertText(t, "     1", formatShortText(1))
	assertText(t, "    10", formatShortText(10))
	assertText(t, "   100", formatShortText(100))
	assertText(t, "  1000", formatShortText(1000))
	assertText(t, " 10000", formatShortText(10000))
	assertText(t, "    -1", formatShortText(-1))
	assertText(t, "   -10", formatShortText(-10))
	assertText(t, "  -100", formatShortText(-100))
	assertText(t, " -1000", formatShortText(-1000))
	assertText(t, "-10000", formatShortText(-10000))
}

func TestWriteIntAsText(t *testing.T) {
	assertText(t, "        1", formatIntText(1))
	assertText(t, "       10", formatIntText(10))
	assertText(t, "      100", formatIntText(100))
	assertText(t, "     1000", formatIntText(1000))
	assertText(t, "    10000", formatIntText(10000))
	assertText(t, "   100000", formatIntText(100000))
	assertText(t, "  1000000", formatIntText(1000000))
	assertText(t, " 10000000", formatIntText(10000000))
	assertText(t, "100000000", formatIntText(100000000))
	assertText(t, "       -1", formatIntText(-1))
	assertText(t, "      -10", formatIntText(-10))
	assertText(t, "     -100", formatIntText(-100))
	assertText(t, "    -1000", formatIntText(-1000))
	assertText(t, "   -10000", formatIntText(-10000))
	assertText(t, "  -100000", formatIntText(-100000))
	assertText(t, " -1000000", formatIntText(-1000000))
	assertText(t, "-10000000", formatIntText(-10000000))
	assertText(t, "-100000000", formatIntText(-100000000))
}

func TestWriteLongAsText(t *testing.T) {
	assertText(t, "1", formatLongText(1))
	assertText(t, "10", formatLongText(10))
	assertText(t, "100", formatLongText(100))
	assertText(t, "-1", formatLongText(-1))
}

func TestWriteDoubleAsText(t *testing.T) {
	assertText(t, "1.0", formatFloat64Text(1.0))
	assertText(t, "1.000005", formatFloat64Text(1.000005))
	assertText(t, "-1.0", formatFloat64Text(-1.0))
	assertText(t, "-1.000005", formatFloat64Text(-1.000005))
	assertText(t, "1500000000000000.0", formatFloat64Text(1.5e15))
}

func TestWriteStringAsText(t *testing.T) {
	assertText(t, "Rep\\U+00E8re pi\\U+00E8ce", formatStringText("Repère pièce", R2004))
	assertText(t, "Repère pièce", formatStringText("Repère pièce", R2007))
}

func TestWriteCodePairAsText(t *testing.T) {
	assertCodePairText(t, "290\r\n     1\r\n", NewBoolCodePair(290, true))
	assertCodePairText(t, " 70\r\n     1\r\n", NewShortCodePair(70, 1))
	assertCodePairText(t, " 90\r\n        1\r\n", NewIntCodePair(90, 1))
	assertCodePairText(t, "160\r\n1\r\n", NewLongCodePair(160, 1))
	assertCodePairText(t, " 10\r\n1.0\r\n", NewDoubleCodePair(10, 1.0))
	assertCodePairText(t, "1010\r\n1.0\r\n", NewDoubleCodePair(1010, 1.0))
	assertCodePairText(t, "  1\r\nabc\r\n", NewStringCodePair(1, "abc"))
}

func TestWriteBoolAsBinary(t *testing.T) {
	assertBinary(t, []byte{0x01, 0x00}, formatBoolBinary(true, R12))
	assertBinary(t, []byte{0x00, 0x00}, formatBoolBinary(false, R12))

	assertBinary(t, []byte{0x01}, formatBoolBinary(true, R13))
	assertBinary(t, []byte{0x00}, formatBoolBinary(false, R13))
}

func TestWriteShortAsBinary(t *testing.T) {
	assertBinary(t, []byte{0x01, 0x00}, formatShortBinary(1))
}

func TestWriteIntAsBinary(t *testing.T) {
	assertBinary(t, []byte{0x01, 0x00, 0x00, 0x00}, formatIntBinary(1))
}

func TestWriteLongAsBinary(t *testing.T) {
	assertBinary(t, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, formatLongBinary(1))
}

func TestWriteDoubleAsBinary(t *testing.T) {
	assertBinary(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}, formatFloat64Binary(1.0))
}

func TestWriteStringAsBinary(t *testing.T) {
	assertBinary(t, []byte{0x61, 0x00}, formatStringBinary("a"))
}

func TestWriteCodePairAsBinary(t *testing.T) {
	assertCodePairBinary(t, []byte{0xFF, 0x22, 0x01, 0x01, 0x00}, NewBoolCodePair(290, true), R12)
	assertCodePairBinary(t, []byte{0x46, 0x01, 0x00}, NewShortCodePair(70, 1), R12)
	assertCodePairBinary(t, []byte{0x5A, 0x01, 0x00, 0x00, 0x00}, NewIntCodePair(90, 1), R12)
	assertCodePairBinary(t, []byte{0xA0, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, NewLongCodePair(160, 1), R12)
	assertCodePairBinary(t, []byte{0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}, NewDoubleCodePair(10, 1.0), R12)
	assertCodePairBinary(t, []byte{0xFF, 0xF2, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}, NewDoubleCodePair(1010, 1.0), R12)
	assertCodePairBinary(t, []byte{0x01, 0x61, 0x00}, NewStringCodePair(1, "a"), R12)

	assertCodePairBinary(t, []byte{0x22, 0x01, 0x01}, NewBoolCodePair(290, true), R13)
	assertCodePairBinary(t, []byte{0x46, 0x00, 0x01, 0x00}, NewShortCodePair(70, 1), R13)
	assertCodePairBinary(t, []byte{0x5A, 0x00, 0x01, 0x00, 0x00, 0x00}, NewIntCodePair(90, 1), R13)
	assertCodePairBinary(t, []byte{0xA0, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, NewLongCodePair(160, 1), R13)
	assertCodePairBinary(t, []byte{0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}, NewDoubleCodePair(10, 1.0), R13)
	assertCodePairBinary(t, []byte{0xF2, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}, NewDoubleCodePair(1010, 1.0), R13)
	assertCodePairBinary(t, []byte{0x01, 0x00, 0x61, 0x00}, NewStringCodePair(1, "a"), R13)
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

func TestWriteBinarySentinelAndStartSection(t *testing.T) {
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

func assertText(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected:\n[%s] but got:\n[%s]", expected, actual)
	}
}

func assertCodePairText(t *testing.T, expected string, codePair CodePair) {
	buf := new(bytes.Buffer)
	writer := newTextCodePairWriter(buf, R14) // version is meaningless here
	err := writer.writeCodePair(codePair)
	if err != nil {
		t.Error(err)
	}

	actual := buf.String()
	if actual != expected {
		t.Errorf("Expected:\n[%s] but got:\n[%s]", expected, actual)
	}
}

func assertBinary(t *testing.T, expected, actual []byte) {
	pass := true
	if len(expected) == len(actual) {
		for i := 0; i < len(expected) && pass; i++ {
			if expected[i] != actual[i] {
				pass = false
				break
			}
		}

		if pass {
			return
		}
	}

	t.Errorf("Expected:\n[% X]\n but found:\n[% X]", expected, actual)
}

func assertCodePairBinary(t *testing.T, expected []byte, codePair CodePair, version AcadVersion) {
	buf := new(bytes.Buffer)
	writer := newBinaryCodePairWriter(buf, version)
	err := writer.writeCodePair(codePair)
	if err != nil {
		t.Error(err)
	}

	actual := buf.Bytes()
	assertBinary(t, expected, actual)
}
