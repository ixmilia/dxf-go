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

	stringValue, ok := codePair.Value.(StringCodePairValue)
	if ok {
		return a.writeString(stringValue.Value)
	}

	return nil
}
