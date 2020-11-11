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
	init() error
	writeCodePair(codePair CodePair) error
}

// code pairs
type directCodePairWriter struct {
	CodePairs []CodePair
}

func newDirectCodePairWriter() directCodePairWriter {
	return directCodePairWriter{
		CodePairs: make([]CodePair, 0),
	}
}

func (d *directCodePairWriter) init() error {
	// noop
	return nil
}

func (d *directCodePairWriter) writeCodePair(codePair CodePair) error {
	d.CodePairs = append(d.CodePairs, codePair)
	return nil
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

func formatShortText(val int16) string {
	return fmt.Sprintf("%6d", val)
}

func formatIntText(val int) string {
	return fmt.Sprintf("%9d", val)
}

func formatLongText(val int64) string {
	return fmt.Sprintf("%d", val)
}

func formatBoolText(val bool) string {
	short := 1
	if !val {
		short = 0
	}

	return formatShortText(int16(short))
}

func formatFloat64Text(val float64) string {
	// trim trailing zeros
	display := strings.TrimRight(fmt.Sprintf("%.12f", val), "0")

	// ensure it doesn't end with a decimal
	if strings.HasSuffix(display, ".") {
		display += "0"
	}

	return display
}

func formatStringText(val string, version AcadVersion) string {
	if version <= R2004 {
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

	return val
}

func (a *textCodePairWriter) writeBoolean(val bool) error {
	return a.writeString(formatBoolText(val))
}

func (a *textCodePairWriter) writeDouble(val float64) error {
	return a.writeString(formatFloat64Text(val))
}

func (a *textCodePairWriter) writeInt(val int) error {
	return a.writeString(formatIntText(val))
}

func (a *textCodePairWriter) writeLong(val int64) error {
	return a.writeString(formatLongText(val))
}

func (a *textCodePairWriter) writeShort(val int16) error {
	return a.writeString(formatShortText(val))
}

func (a *textCodePairWriter) writeString(val string) error {
	formatted := formatStringText(val, a.version)
	bytes := []byte(fmt.Sprintf("%s\r\n", formatted))
	_, err := a.writer.Write(bytes)
	return err
}

func (a *textCodePairWriter) init() error {
	// noop
	return nil
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

func newBinaryCodePairWriter(writer io.Writer, version AcadVersion) codePairWriter {
	return &binaryCodePairWriter{
		writer:  writer,
		version: version,
	}
}

func formatShortBinary(val int16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(val))
	return buf
}

func formatIntBinary(val int) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(val))
	return buf
}

func formatLongBinary(val int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(val))
	return buf
}

func formatBoolBinary(val bool, version AcadVersion) []byte {
	short := 1
	if !val {
		short = 0
	}

	// after R13 bools are a single byte
	if version >= R13 {
		return []byte{byte(short)}
	}

	return formatShortBinary(int16(short))
}

func formatFloat64Binary(val float64) []byte {
	buf := make([]byte, 8)
	ui := math.Float64bits(val)
	binary.LittleEndian.PutUint64(buf, ui)
	return buf
}

func formatStringBinary(val string) []byte {
	buf := []byte(val)
	buf = append(buf, 0x00)
	return buf
}

func (b *binaryCodePairWriter) init() error {
	sentinel := []byte("AutoCAD Binary DXF\r\n")
	sentinel = append(sentinel, []byte{0x1A, 0x00}...)
	n, err := b.writer.Write(sentinel)
	if err != nil {
		return err
	}

	if n != len(sentinel) {
		return errors.New("unable to write binary sentinel")
	}

	return nil
}

func (b *binaryCodePairWriter) writeCodePair(codePair CodePair) error {
	var err error
	if b.version >= R13 {
		// after R13 codes are always 2 bytes
		err = b.writeShort(int16(codePair.Code))
		if err != nil {
			return err
		}
	} else if codePair.Code >= 255 {
		// before R13 codes were 1 or 3 bytes
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
		return b.writeBoolean(t.Value, b.version)
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

func (b *binaryCodePairWriter) writeShort(val int16) error {
	return b.writeBytes(formatShortBinary(val))
}

func (b *binaryCodePairWriter) writeBoolean(val bool, version AcadVersion) error {
	return b.writeBytes(formatBoolBinary(val, version))
}

func (b *binaryCodePairWriter) writeInt(val int) error {
	return b.writeBytes(formatIntBinary(val))
}

func (b *binaryCodePairWriter) writeLong(val int64) error {
	return b.writeBytes(formatLongBinary(val))
}

func (b *binaryCodePairWriter) writeDouble(val float64) error {
	return b.writeBytes(formatFloat64Binary(val))
}

func (b *binaryCodePairWriter) writeString(val string) error {
	return b.writeBytes(formatStringBinary(val))
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
