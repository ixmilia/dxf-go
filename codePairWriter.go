package dxf

import (
	"fmt"
	"io"
	"strings"
)

type CodePairWriter interface {
	writeCodePair(codePair CodePair) error
}

// ASCII
type AsciiCodePairWriter struct {
	writer io.Writer
}

func NewAsciiCodePairWriter(writer io.Writer) CodePairWriter {
	return AsciiCodePairWriter{
		writer: writer,
	}
}

func (a AsciiCodePairWriter) writeDouble(val float64) error {
	// trim trailing zeros
	display := strings.TrimRight(fmt.Sprintf("%.12f", val), "0")

	// ensure it doesn't end with a decimal
	if strings.HasSuffix(display, ".") {
		display += "0"
	}

	return a.writeString(display)
}

func (a AsciiCodePairWriter) writeShort(val int16) error {
	return a.writeString(fmt.Sprintf("%6d", val))
}

func (a AsciiCodePairWriter) writeString(val string) error {
	bytes := []byte(fmt.Sprintf("%s\r\n", val))
	_, err := a.writer.Write(bytes)
	return err
}

func (a AsciiCodePairWriter) writeCodePair(codePair CodePair) error {
	err := a.writeString(fmt.Sprintf("%3d", codePair.Code))
	if err != nil {
		return nil
	}

	switch t := codePair.Value.(type) {
	case DoubleCodePairValue:
		return a.writeDouble(t.Value)
	case ShortCodePairValue:
		return a.writeShort(t.Value)
	case StringCodePairValue:
		return a.writeString(t.Value)
	}

	return nil
}
