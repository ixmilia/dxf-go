package dxf

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type codePairReader interface {
	readCodePair() (CodePair, error)
	setUtf8Reader()
}

func codePairReaderFromReader(reader io.Reader, e encoding.Encoding) (r codePairReader, err error) {
	decoder := *e.NewDecoder()
	firstLine, err := readSingleLine(reader, decoder)
	if err != nil {
		r = newDirectCodePairReader()
		if firstLine == "" {
			// empty file is valid
			err = nil
			return
		}

		return
	}

	if firstLine == "AutoCAD Binary DXF" {
		r, err = newBinaryCodePairReader(reader)
	} else {
		r = newTextCodePairReader(reader, decoder, firstLine)
	}

	return r, err
}

// code pairs
type directCodePairReader struct {
	index     int
	codePairs []CodePair
}

func newDirectCodePairReader(codePairs ...CodePair) codePairReader {
	return &directCodePairReader{
		index:     0,
		codePairs: codePairs,
	}
}

func (d *directCodePairReader) readCodePair() (codePair CodePair, err error) {
	if d.index >= len(d.codePairs) {
		err = errors.New("out of data")
	} else {
		codePair = d.codePairs[d.index]
		d.index++
	}

	return codePair, err
}

func (d *directCodePairReader) setUtf8Reader() {
	// noop
}

// text
type textCodePairReader struct {
	reader        io.Reader
	decoder       encoding.Decoder
	firstLine     string
	firstLineRead bool
	readAsUtf8    bool
}

func newTextCodePairReader(reader io.Reader, decoder encoding.Decoder, firstLine string) codePairReader {
	return &textCodePairReader{
		reader:        reader,
		decoder:       decoder,
		firstLine:     firstLine,
		firstLineRead: false,
		readAsUtf8:    false,
	}
}

func readSingleLine(reader io.Reader, d encoding.Decoder) (line string, err error) {
	buffer := make([]byte, 1)
	bytes := make([]byte, 0)

	for {
		count, e := reader.Read(buffer)
		if e == io.EOF {
			e = nil
		}
		if e != nil {
			err = e
			return
		}
		if count != 1 {
			break
		}
		if buffer[0] == '\n' {
			break
		}

		bytes = append(bytes, buffer[0])
	}

	line, _, err = transform.String(d.Transformer, string(bytes))
	if err != nil {
		return
	}

	if strings.HasSuffix(line, "\r") {
		line = line[:len(line)-1]
	}

	return
}

func (a *textCodePairReader) readLine(d encoding.Decoder) (line string, err error) {
	if !a.firstLineRead {
		line = a.firstLine
		a.firstLine = ""
		a.firstLineRead = true
		return
	}

	line, err = readSingleLine(a.reader, d)
	return
}

func (a *textCodePairReader) readCode() (int, error) {
	line, err := a.readLine(*encoding.Nop.NewDecoder())
	if err != nil {
		return 0, err
	}

	code, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, err
	}

	return code, nil
}

func readBoolText(line string) (bool, error) {
	value, err := readShortText(line)
	result := value != 0
	return result, err
}

func readShortText(line string) (int16, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(line), 10, 16)
	result := int16(value)
	return result, err
}

func readIntText(line string) (int, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(line), 10, 32)
	result := int(value)
	return result, err
}

func readLongText(line string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(line), 10, 64)
}

func readDoubleText(line string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(line), 64)
}

func readStringText(line string, readAsUtf8 bool) (value string, err error) {
	if !readAsUtf8 {
		line = parseUtf8(line)
	}

	return line, nil
}

func (a *textCodePairReader) readCodePair() (CodePair, error) {
	var codePair CodePair
	code, err := a.readCode()
	if err != nil {
		return codePair, err
	}

	stringValue, err := a.readLine(a.decoder)
	if err != nil {
		return codePair, err
	}

	switch codeTypeName(code) {
	case "Bool":
		value, err := readBoolText(stringValue)
		if err != nil {
			return codePair, err
		}
		codePair = NewBoolCodePair(code, value)
	case "Double":
		value, err := readDoubleText(stringValue)
		if err != nil {
			return codePair, err
		}
		codePair = NewDoubleCodePair(code, value)
	case "Int":
		value, err := readIntText(stringValue)
		if err != nil {
			return codePair, err
		}
		codePair = NewIntCodePair(code, value)
	case "Long":
		value, err := readLongText(stringValue)
		if err != nil {
			return codePair, err
		}
		codePair = NewLongCodePair(code, value)
	case "Short":
		value, err := readShortText(stringValue)
		if err != nil {
			return codePair, err
		}
		codePair = NewShortCodePair(code, value)
	case "String":
		value, err := readStringText(stringValue, a.readAsUtf8)
		if err != nil {
			return codePair, err
		}
		codePair = NewStringCodePair(code, value)
	}

	return codePair, nil
}

func (a *textCodePairReader) setUtf8Reader() {
	a.decoder = *unicode.UTF8.NewDecoder()
	a.readAsUtf8 = true
}

// binary
type binaryCodePairReader struct {
	reader          bufio.Reader
	hasReturnedPair bool
	isPostR13       bool
}

func newBinaryCodePairReader(reader io.Reader) (rdr codePairReader, err error) {
	r := *bufio.NewReader(reader)
	buf := make([]byte, 2)
	n, err := r.Read(buf)
	if err != nil {
		return
	}
	if n != 2 {
		err = errors.New("not enough bytes")
		return
	}
	if buf[0] != 0x1A || buf[1] != 0x00 {
		err = errors.New("expected 0x1A, 0x00")
		return
	}
	rdr = &binaryCodePairReader{
		reader:          r,
		hasReturnedPair: false,
		isPostR13:       false,
	}
	return
}

func readBytes(reader *bufio.Reader, count int) (buf []byte, err error) {
	buf = make([]byte, count)
	n, err := reader.Read(buf)
	if err != nil {
		return
	}

	if n != len(buf) {
		err = errors.New("not enough bytes")
		return
	}

	return
}

func readBoolBinary(data []byte, isPostR13 bool) (val bool, err error) {
	// after R13 bools are encoded as a single byte
	if isPostR13 {
		if len(data) != 1 {
			err = errors.New("Expected 1 byte to read post R13 bool.")
			return
		}

		val = data[0] != 0
	} else {
		if len(data) != 2 {
			err = errors.New("Expected 2 bytes to read pre R13 bool.")
			return
		}

		s := createShort(data[0], data[1])
		val = s != 0
	}

	return
}

func readShortBinary(data []byte) (val int16, err error) {
	if len(data) != 2 {
		err = errors.New("Expected 2 bytes to read int16.")
		return
	}

	val = createShort(data[0], data[1])
	return
}

func readIntBinary(data []byte) (val int, err error) {
	if len(data) != 4 {
		err = errors.New("Expected 4 bytes to read int.")
		return
	}

	uval := binary.LittleEndian.Uint32(data)
	val = int(uval)
	return
}

func readLongBinary(data []byte) (val int64, err error) {
	if len(data) != 8 {
		err = errors.New("Expected 8 bytes to read int64.")
		return
	}

	uval := binary.LittleEndian.Uint64(data)
	val = int64(uval)
	return
}

func readDoubleBinary(data []byte) (val float64, err error) {
	if len(data) != 8 {
		err = errors.New("Expected 8 bytes to read float64.")
		return
	}

	uval := binary.LittleEndian.Uint64(data)
	val = math.Float64frombits(uval)
	return
}

func readStringBinary(reader *bufio.Reader) (val string, err error) {
	buf := make([]byte, 0)
	for {
		var c byte
		c, err = reader.ReadByte()
		if err != nil {
			return
		}
		if c == 0x00 {
			break
		}
		buf = append(buf, c)
	}

	val = string(buf)
	return
}

func (b *binaryCodePairReader) readCodePair() (CodePair, error) {
	var pair CodePair
	var err error
	code, err := b.readCode()
	if err != nil {
		return pair, err
	}

	switch codeTypeName(code) {
	case "Bool":
		boolByteCount := 2
		if b.isPostR13 {
			boolByteCount = 1
		}
		data, err := readBytes(&b.reader, boolByteCount)
		if err != nil {
			return pair, err
		}
		value, err := readBoolBinary(data, b.isPostR13)
		if err != nil {
			return pair, err
		}
		pair = NewBoolCodePair(code, value)
	case "Double":
		data, err := readBytes(&b.reader, 8)
		if err != nil {
			return pair, err
		}
		value, err := readDoubleBinary(data)
		if err != nil {
			return pair, err
		}
		pair = NewDoubleCodePair(code, value)
	case "Int":
		data, err := readBytes(&b.reader, 4)
		if err != nil {
			return pair, err
		}
		value, err := readIntBinary(data)
		if err != nil {
			return pair, err
		}
		pair = NewIntCodePair(code, value)
	case "Long":
		data, err := readBytes(&b.reader, 8)
		if err != nil {
			return pair, err
		}
		value, err := readLongBinary(data)
		if err != nil {
			return pair, err
		}
		pair = NewLongCodePair(code, value)
	case "Short":
		data, err := readBytes(&b.reader, 2)
		if err != nil {
			return pair, err
		}
		value, err := readShortBinary(data)
		if err != nil {
			return pair, err
		}
		pair = NewShortCodePair(code, value)
	case "String":
		value, err := readStringBinary(&b.reader)
		if err != nil {
			return pair, err
		}
		pair = NewStringCodePair(code, value)
	}

	return pair, err
}

func (b *binaryCodePairReader) readCode() (code int, err error) {
	bt, err := b.readByte()
	if err != nil {
		return
	}
	code = int(bt)

	if !b.hasReturnedPair && code == 0 {
		p := make([]byte, 1)
		p, err = b.reader.Peek(1)
		if err != nil {
			return
		}
		if p[0] == 0 {
			// The first code pair in a binary file must be `0/SECTION`; if we're reading the first pair, the code is
			// `0`, and the next byte is NULL (empty string), then this must be a post R13 file where codes are always
			// encoded with 2 bytes.
			b.isPostR13 = true
		}
	}

	// potentially read the second byte of the code
	if b.isPostR13 {
		var b2 byte
		b2, err = b.readByte()
		if err != nil {
			return
		}
		code = int(createShort(bt, b2))
	} else if code == 255 {
		var data []byte
		data, err = readBytes(&b.reader, 2)
		if err != nil {
			return
		}
		var s int16
		s, err = readShortBinary(data)
		if err != nil {
			return
		}
		code = int(s)
	}

	b.hasReturnedPair = true
	return
}

func (b *binaryCodePairReader) readByte() (byte, error) {
	return b.reader.ReadByte()
}

func createShort(b1, b2 byte) int16 {
	return int16(b2)<<8 + int16(b1)
}

func (b *binaryCodePairReader) setUtf8Reader() {
	// noop
}

func parseUtf8(v string) string {
	var final strings.Builder
	var seq strings.Builder
	inEscapeSequence := false
	sequenceStart := 0
	for i, r := range v {
		if !inEscapeSequence {
			if r == '\\' {
				inEscapeSequence = true
				sequenceStart = i
				seq.Reset()
				seq.WriteRune(r)
			} else {
				final.WriteRune(r)
			}
		} else {
			seq.WriteRune(r)
			if i == sequenceStart+6 {
				inEscapeSequence = false
				escaped := seq.String()
				seq.Reset()
				if strings.HasPrefix(escaped, "\\U+") {
					codeStr := escaped[3:]
					code, err := strconv.ParseUint(codeStr, 16, 64)
					if err == nil {
						final.WriteRune(rune(code))
					} else {
						final.WriteRune('?')
					}
				} else {
					final.WriteString(escaped)
				}
			}
		}
	}

	final.WriteString(seq.String())
	return final.String()
}
