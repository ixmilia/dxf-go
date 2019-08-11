package dxf

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

type codePairReader interface {
	readCodePair() (CodePair, error)
	setUtf8Reader()
}

// text
type textCodePairReader struct {
	scanner    bufio.Scanner
	readAsUtf8 bool
}

func newTextCodePairReader(reader io.Reader) codePairReader {
	return &textCodePairReader{
		scanner:    *bufio.NewScanner(reader),
		readAsUtf8: false,
	}
}

func (a *textCodePairReader) readLine() (line string, err error) {
	if !a.scanner.Scan() {
		err = a.scanner.Err()
		return
	}
	line = a.scanner.Text()
	return
}

func (a *textCodePairReader) readCode() (int, error) {
	line, err := a.readLine()
	if err != nil {
		return 0, err
	}

	code, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (a *textCodePairReader) readCodePair() (CodePair, error) {
	var codePair CodePair
	code, err := a.readCode()
	if err != nil {
		return codePair, err
	}

	stringValue, err := a.readLine()
	if err != nil {
		return codePair, err
	}

	switch codeTypeName(code) {
	case "Bool":
		value, err := strconv.ParseInt(strings.TrimSpace(stringValue), 10, 16)
		if err != nil {
			return codePair, err
		}
		codePair = NewBoolCodePair(code, value != 0)
	case "Double":
		value, err := strconv.ParseFloat(strings.TrimSpace(stringValue), 64)
		if err != nil {
			return codePair, err
		}
		codePair = NewDoubleCodePair(code, value)
	case "Int":
		value, err := strconv.ParseInt(strings.TrimSpace(stringValue), 10, 32)
		if err != nil {
			return codePair, err
		}
		codePair = NewIntCodePair(code, int(value))
	case "Long":
		value, err := strconv.ParseInt(strings.TrimSpace(stringValue), 10, 64)
		if err != nil {
			return codePair, err
		}
		codePair = NewLongCodePair(code, value)
	case "Short":
		value, err := strconv.ParseInt(strings.TrimSpace(stringValue), 10, 16)
		if err != nil {
			return codePair, err
		}
		codePair = NewShortCodePair(code, int16(value))
	case "String":
		if !a.readAsUtf8 {
			stringValue = parseUtf8(stringValue)
		}
		codePair = NewStringCodePair(code, stringValue)
	}

	return codePair, nil
}

func (a *textCodePairReader) setUtf8Reader() {
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

func (b *binaryCodePairReader) readCodePair() (CodePair, error) {
	var pair CodePair
	var err error
	code, err := b.readCode()
	if err != nil {
		return pair, err
	}

	switch codeTypeName(code) {
	case "Bool":
		value, err := b.readBool()
		if err != nil {
			return pair, err
		}
		pair = NewBoolCodePair(code, value)
	case "Double":
		value, err := b.readDouble()
		if err != nil {
			return pair, err
		}
		pair = NewDoubleCodePair(code, value)
	case "Int":
		value, err := b.readInt()
		if err != nil {
			return pair, err
		}
		pair = NewIntCodePair(code, int(value))
	case "Long":
		value, err := b.readLong()
		if err != nil {
			return pair, err
		}
		pair = NewLongCodePair(code, value)
	case "Short":
		value, err := b.readShort()
		if err != nil {
			return pair, err
		}
		pair = NewShortCodePair(code, int16(value))
	case "String":
		buf := make([]byte, 0)
		for {
			c, err := b.readByte()
			if err != nil {
				return pair, err
			}
			if c == 0x00 {
				break
			}
			buf = append(buf, c)
		}
		value := string(buf)
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
		var s int16
		s, err = b.readShort()
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

func (b *binaryCodePairReader) readBool() (r bool, err error) {
	if b.isPostR13 {
		// after R13 bools are encoded as a single byte
		var t byte
		t, err = b.readByte()
		if err != nil {
			return
		}

		r = t != 0
	} else {
		var v int16
		v, err = b.readShort()
		if err != nil {
			return
		}

		r = v != 0
	}
	return
}

func (b *binaryCodePairReader) readShort() (s int16, err error) {
	b1, err := b.readByte()
	if err != nil {
		return
	}
	b2, err := b.readByte()
	if err != nil {
		return
	}
	s = createShort(b1, b2)
	return
}

func (b *binaryCodePairReader) readInt() (i int, err error) {
	buf := make([]byte, 4)
	n, err := b.reader.Read(buf)
	if err != nil {
		return
	}
	if n != len(buf) {
		err = errors.New("not enough bytes")
		return
	}
	u := binary.LittleEndian.Uint32(buf)
	i = int(u)
	return i, err
}

func (b *binaryCodePairReader) readLong() (l int64, err error) {
	buf := make([]byte, 8)
	n, err := b.reader.Read(buf)
	if err != nil {
		return
	}
	if n != len(buf) {
		err = errors.New("not enough bytes")
		return
	}
	u := binary.LittleEndian.Uint64(buf)
	l = int64(u)
	return l, err
}

func (b *binaryCodePairReader) readDouble() (d float64, err error) {
	buf := make([]byte, 8)
	n, err := b.reader.Read(buf)
	if err != nil {
		return
	}
	if n != len(buf) {
		err = errors.New("not enough bytes")
		return
	}
	u := binary.LittleEndian.Uint64(buf)
	d = math.Float64frombits(u)
	return d, err
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
