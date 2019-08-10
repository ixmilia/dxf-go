package dxf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

type codePairWriter interface {
	writeCodePair(codePair CodePair) error
}

// text
type textCodePairWriter struct {
	writer  io.Writer
	version AcadVersion
}

func newTextCodePairWriter(writer io.Writer, version AcadVersion) codePairWriter {
	return &textCodePairWriter{
		writer:  writer,
		version: version,
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

func (a *textCodePairWriter) writeBoolean(val bool) error {
	short := 1
	if !val {
		short = 0
	}
	return a.writeShort(int16(short))
}

func (a *textCodePairWriter) writeDouble(val float64) error {
	return a.writeString(formatFloat64(val))
}

func (a *textCodePairWriter) writeInt(val int) error {
	return a.writeString(fmt.Sprintf("%9d", val))
}

func (a *textCodePairWriter) writeLong(val int64) error {
	return a.writeString(fmt.Sprintf("%d", val))
}

func (a *textCodePairWriter) writeShort(val int16) error {
	return a.writeString(fmt.Sprintf("%6d", val))
}

func (a *textCodePairWriter) writeString(val string) error {
	if a.version <= R2004 {
		// escape unicode characters
		var builder strings.Builder
		for _, r := range val {
			u := uint(r)
			if u >= 128 {
				builder.WriteString(fmt.Sprintf("\\U+%04X", u))
			} else {
				builder.WriteRune(r)
			}
		}

		val = builder.String()
	}

	bytes := []byte(fmt.Sprintf("%s\r\n", val))
	_, err := a.writer.Write(bytes)
	return err
}

func (a *textCodePairWriter) writeCodePair(codePair CodePair) error {
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

// binary
type binaryCodePairWriter struct {
	writer  io.Writer
	version AcadVersion
}

func newBinaryCodePairWriter(writer io.Writer, version AcadVersion) (codePairWriter, error) {
	var w codePairWriter
	var err error
	sentinel := []byte("AutoCAD Binary DXF\r\n")
	sentinel = append(sentinel, []byte{0x1A, 0x00}...)
	n, err := writer.Write(sentinel)
	if err != nil {
		return nil, err
	}
	if n != len(sentinel) {
		return nil, errors.New("unable to write binary sentinel")
	}

	w = &binaryCodePairWriter{
		writer:  writer,
		version: version,
	}

	return w, nil
}

func (b *binaryCodePairWriter) writeCodePair(codePair CodePair) error {
	var err error
	if codePair.Code >= 255 {
		err = b.writeByte(255)
		if err != nil {
			return err
		}
		err = b.writeShort(int16(codePair.Code))
	} else {
		err = b.writeByte(byte(codePair.Code))
	}
	if err != nil {
		return err
	}

	switch t := codePair.Value.(type) {
	case BoolCodePairValue:
		return b.writeBoolean(t.Value)
	case DoubleCodePairValue:
		return b.writeDouble(t.Value)
	case IntCodePairValue:
		return b.writeInt(t.Value)
	case LongCodePairValue:
		return b.writeLong(t.Value)
	case ShortCodePairValue:
		return b.writeShort(t.Value)
	case StringCodePairValue:
		return b.writeString(t.Value)
	default:
		return fmt.Errorf("unsupported code pair value type %T", t)
	}
}

func (b *binaryCodePairWriter) writeBytes(buf []byte) error {
	n, err := b.writer.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errors.New("unable to write bytes")
	}

	return nil
}

func (b *binaryCodePairWriter) writeByte(bt byte) error {
	return b.writeBytes([]byte{bt})
}

func (b *binaryCodePairWriter) writeShort(s int16) error {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(s))
	return b.writeBytes(buf)
}

func (b *binaryCodePairWriter) writeBoolean(v bool) error {
	var s int16
	if v {
		s = 1
	} else {
		s = 0
	}
	return b.writeShort(s)
}

func (b *binaryCodePairWriter) writeInt(v int) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	return b.writeBytes(buf)
}

func (b *binaryCodePairWriter) writeLong(v int64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(v))
	return b.writeBytes(buf)
}

func (b *binaryCodePairWriter) writeDouble(v float64) error {
	buf := make([]byte, 8)
	u := math.Float64bits(v)
	binary.LittleEndian.PutUint64(buf, u)
	return b.writeBytes(buf)
}

func (b *binaryCodePairWriter) writeString(v string) error {
	buf := []byte(v)
	err := b.writeBytes(buf)
	if err != nil {
		return err
	}
	return b.writeByte(0x00)
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
