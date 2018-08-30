package dxf

import (
	"bytes"
	"os"
)

type Drawing struct {
	Header Header
}

func NewDrawing() *Drawing {
	return &Drawing{
		Header: *NewHeader(),
	}
}

func (d Drawing) SaveFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	writer := NewAsciiCodePairWriter(f)
	return d.saveToWriter(writer)
}

func (d Drawing) saveToWriter(writer CodePairWriter) error {
	err := d.Header.writeHeader(writer)
	if err != nil {
		return err
	}

	err = writer.writeCodePair(NewStringCodePair(0, "EOF"))
	return err
}

func (d Drawing) String() string {
	buf := new(bytes.Buffer)
	writer := NewAsciiCodePairWriter(buf)
	d.saveToWriter(writer)
	return buf.String()
}
