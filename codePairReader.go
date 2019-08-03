package dxf

import (
	"bufio"
	"io"
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
