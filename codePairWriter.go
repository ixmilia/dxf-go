package dxf

import (
	"fmt"
	"io"
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
	case ShortCodePairValue:
		return a.writeShort(t.Value)
	case StringCodePairValue:
		return a.writeString(t.Value)
	}

	return nil
}
