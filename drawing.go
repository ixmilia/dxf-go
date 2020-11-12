package dxf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/text/encoding"
)

// The Drawing struct represents a complete DXF drawing.
type Drawing struct {
	Header Header

	AppIds       []AppId
	BlockRecords []BlockRecord
	DimStyles    []DimStyle
	Layers       []Layer
	LineTypes    []LineType
	Styles       []Style
	Ucss         []Ucs
	Views        []View
	ViewPorts    []ViewPort

	Entities []Entity
}

// NewDrawing returns a new, fully initialized drawing.
func NewDrawing() *Drawing {
	return &Drawing{
		Header:   *NewHeader(),
		Entities: make([]Entity, 0),
	}
}

// GetItemByHandle gets a `DrawingItem` with the appropriate handle.
func (d *Drawing) GetItemByHandle(h Handle) (item *DrawingItem, err error) {
	item = nil
	err = nil

	for i := range d.Entities {
		e := &d.Entities[i]
		if (*e).Handle() == h {
			di := (*e).(DrawingItem)
			item = &di
			return
		}
	}

	err = fmt.Errorf("Unable to find item with handle '%d'", h)
	return
}

// SaveFile writes the current drawing to the specified path.
func (d *Drawing) SaveFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	return d.SaveToWriter(f)
}

// SaveFileBinary writes the current drawing to the specified path as a binary DXF.
func (d *Drawing) SaveFileBinary(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	return d.SaveToWriterBinary(f)
}

// SaveToWriter writes the current drawing to the specified io.Writer.
func (d *Drawing) SaveToWriter(writer io.Writer) error {
	codePairWriter := newTextCodePairWriter(writer, d.Header.Version)
	return d.saveToCodePairWriter(codePairWriter)
}

// SaveToWriterBinary writes the current drawing to the specified io.Writer as a binary DXF.
func (d *Drawing) SaveToWriterBinary(writer io.Writer) error {
	codePairWriter := newBinaryCodePairWriter(writer, d.Header.Version)
	return d.saveToCodePairWriter(codePairWriter)
}

func (d *Drawing) String() string {
	buf := new(bytes.Buffer)
	err := d.SaveToWriter(buf)
	if err != nil {
		return err.Error()
	}

	return buf.String()
}

// CodePairs returns the series of `CodePair` that represents the drawing.
func (d *Drawing) CodePairs() (codePairs []CodePair, err error) {
	writer := newDirectCodePairWriter()
	err = d.saveToCodePairWriter(&writer)
	if err != nil {
		return
	}

	codePairs = writer.CodePairs
	return
}

func (d *Drawing) saveToCodePairWriter(writer codePairWriter) error {
	err := writer.init()
	if err != nil {
		return err
	}

	assignHandles(d)
	assignPointers(d)

	err = d.Header.writeHeaderSection(writer)
	if err != nil {
		return err
	}

	err = writeTablesSection(d, writer, d.Header.Version)
	if err != nil {
		return err
	}

	err = writeEntitiesSection(d.Entities, writer, d.Header.Version)
	if err != nil {
		return err
	}

	err = writer.writeCodePair(NewStringCodePair(0, "EOF"))
	return err
}

// ReadFile reads a DXF drawing from the specified path.
func ReadFile(path string) (Drawing, error) {
	var drawing Drawing
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return drawing, err
	}

	return ReadFromReader(bytes.NewReader(buf))
}

// ReadFromReader reads a DXF drawing from the specified io.Reader.
func ReadFromReader(reader io.Reader) (drawing Drawing, err error) {
	return ReadFromReaderWithEncoding(reader, encoding.Nop)
}

// ReadFromReaderWithEncoding reads a DXF drawing from the specified io.Reader with the specified default text encoding.
func ReadFromReaderWithEncoding(reader io.Reader, e encoding.Encoding) (drawing Drawing, err error) {
	r, err := codePairReaderFromReader(reader, e)
	drawing, err = readFromCodePairReader(r)
	return
}

// ParseDrawing returns a drawing as parsed from a `string`.
func ParseDrawing(content string) (Drawing, error) {
	stringReader := strings.NewReader(content)
	return ReadFromReader(stringReader)
}

// ParseDrawingFromCodePairs returns a drawing as parsed from a sequence of `CodePair`.
func ParseDrawingFromCodePairs(codePairs ...CodePair) (Drawing, error) {
	directReader := newDirectCodePairReader(codePairs...)
	return readFromCodePairReader(directReader)
}

func readFromCodePairReader(reader codePairReader) (Drawing, error) {
	drawing := *NewDrawing()

	// read sections
	nextPair, err := reader.readCodePair()

	// parse sections
	for err == nil && !nextPair.isEOF() {
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

		sectionType := nextPair.Value.(StringCodePairValue).Value
		nextPair, err = reader.readCodePair()
		for err == nil && !nextPair.isEndSection() {
			switch sectionType {
			case "ENTITIES":
				drawing.Entities, nextPair, err = readEntities(nextPair, reader)
			case "HEADER":
				drawing.Header, nextPair, err = readHeader(nextPair, reader)
			case "TABLES":
				nextPair, err = readTables(&drawing, nextPair, reader)
			default:
				// swallow unsupported section
				for err == nil && !nextPair.isEndSection() {
					nextPair, err = reader.readCodePair()
				}
			}
		}

		// find 0/ENDSEC
		if err != nil {
			return drawing, err
		}
		if !nextPair.isEndSection() {
			return drawing, errors.New("expected 0/ENDSEC")
		}

		nextPair, err = reader.readCodePair()
	}

	// find possible 0/EOF
	if err != nil {
		// don't care at this point, the file could be done
		err = nil
	} else if !nextPair.isEOF() {
		return drawing, errors.New("expected 0/EOF")
	}

	bindPointers(&drawing)
	return drawing, nil
}

func assignHandles(d *Drawing) {
	nextHandle := uint32(1)
	for i := range d.Entities {
		e := &d.Entities[i]
		if (*e).Handle() == 0 {
			(*e).SetHandle(Handle(nextHandle))
			nextHandle++
		}
	}

	d.Header.NextAvailableHandle = Handle(nextHandle)
}

func assignPointers(d *Drawing) {
	for i := range d.Entities {
		e := &d.Entities[i]
		for _, p := range (*e).pointers() {
			if p.handle == 0 && p.value != nil {
				p.handle = (*p.value).Handle()
			}
		}
	}
}

func bindPointers(d *Drawing) {
	for i := range d.Entities {
		e := &d.Entities[i]
		for _, p := range (*e).pointers() {
			if p.handle != 0 {
				o, err := d.GetItemByHandle(p.handle)
				if err == nil {
					p.value = o
				}
			}
		}
	}
}
