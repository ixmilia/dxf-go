package dxf

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
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

func ReadFile(path string) (Drawing, error) {
	var drawing Drawing
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return drawing, err
	}

	reader := NewAsciiCodePairReader(bytes.NewReader(buf))
	return readFromReader(reader)
}

func readFromReader(reader CodePairReader) (Drawing, error) {
	drawing := *NewDrawing()

	// read sections

	// find 0/SECTION
	nextPair, err := reader.readCodePair()
	if err != nil {
		if err == io.EOF {
			return drawing, nil
		} else {
			return drawing, err
		}
	}

	if !nextPair.isStartSection() {
		return drawing, errors.New("expected 0/SECTION code pair")
	}

	// find 2/<section-type>
	nextPair, err = reader.readCodePair()
	if err != nil {
		return drawing, err
	}
	if nextPair.Code != 2 {
		return drawing, errors.New("expected 2/<section-type>")
	}

	// parse section
	sectionType := nextPair.Value.(StringCodePairValue).Value
	nextPair, err = reader.readCodePair()
	for err == nil && !nextPair.isEndSection() {
		switch sectionType {
		case "HEADER":
			drawing.Header, nextPair, err = readHeader(nextPair, reader)
		}
	}

	// find 0/ENDSEC
	if err != nil {
		return drawing, err
	}
	if !nextPair.isEndSection() {
		return drawing, errors.New("expected 0/ENDSEC")
	}

	// find possible 0/EOF
	nextPair, err = reader.readCodePair()
	if err != nil {
		// don't care at this point, the file could be done
		return drawing, nil
	}
	if !nextPair.isEof() {
		return drawing, errors.New("expected 0/EOF")
	}

	return drawing, nil
}

func ParseDrawing(content string) (Drawing, error) {
	stringReader := strings.NewReader(content)
	reader := NewAsciiCodePairReader(stringReader)
	return readFromReader(reader)
}
