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

func (a asciiCodePairWriter) writeDouble(val float64) error {
	return a.writeString(formatFloat64(val))
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
	case DoubleCodePairValue:
		return a.writeDouble(t.Value)
	case ShortCodePairValue:
		return a.writeShort(t.Value)
	case StringCodePairValue:
		return a.writeString(t.Value)
	}

	return nil
}
