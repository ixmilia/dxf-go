package dxf

import (
	"fmt"
	"io"
	"strings"
)

type codePairWriter interface {
	writeCodePair(codePair CodePair) error
}

// ASCII
type asciiCodePairWriter struct {
	writer io.Writer
}

func newASCIICodePairWriter(writer io.Writer) codePairWriter {
	return asciiCodePairWriter{
		writer: writer,
	}
}

func formatFloat64(val float64) string {
	// trim trailing zeros
	display := strings.TrimRight(fmt.Sprintf("%.12f", val), "0")

	// ensure it doesn't end with a decimal
	if strings.HasSuffix(display, ".") {
		display += "0"
	}

	return display
}

func (a asciiCodePairWriter) writeBoolean(val bool) error {
	short := 1
	if !val {
		short = 0
	}
	return a.writeShort(int16(short))
}

func (a asciiCodePairWriter) writeDouble(val float64) error {
	return a.writeString(formatFloat64(val))
}

func (a asciiCodePairWriter) writeInt(val int) error {
	return a.writeString(fmt.Sprintf("%9d", val))
}

func (a asciiCodePairWriter) writeLong(val int64) error {
	return a.writeString(fmt.Sprintf("%d", val))
}

func (a asciiCodePairWriter) writeShort(val int16) error {
	return a.writeString(fmt.Sprintf("%6d", val))
}

func (a asciiCodePairWriter) writeString(val string) error {
	bytes := []byte(fmt.Sprintf("%s\r\n", val))
	_, err := a.writer.Write(bytes)
	return err
}

func (a asciiCodePairWriter) writeCodePair(codePair CodePair) error {
	err := a.writeString(fmt.Sprintf("%3d", codePair.Code))
	if err != nil {
		return nil
	}

	switch t := codePair.Value.(type) {
	case BoolCodePairValue:
		return a.writeBoolean(t.Value)
	case DoubleCodePairValue:
		return a.writeDouble(t.Value)
	case IntCodePairValue:
		return a.writeInt(t.Value)
	case LongCodePairValue:
		return a.writeLong(t.Value)
	case ShortCodePairValue:
		return a.writeShort(t.Value)
	case StringCodePairValue:
		return a.writeString(t.Value)
	default:
		return fmt.Errorf("unsupported code pair value type %T", t)
	}
}

func writeSectionStart(writer codePairWriter, sectionName string) (error error) {
	error = writer.writeCodePair(NewStringCodePair(0, "SECTION"))
	if error != nil {
		return
	}

	error = writer.writeCodePair(NewStringCodePair(2, sectionName))
	return
}

func writeSectionEnd(writer codePairWriter) (error error) {
	error = writer.writeCodePair(NewStringCodePair(0, "ENDSEC"))
	return
}
