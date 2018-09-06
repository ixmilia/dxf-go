package dxf

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type codePairReader interface {
	readCodePair() (CodePair, error)
}

// ASCII
type asciiCodePairReader struct {
	scanner bufio.Scanner
}

func newASCIICodePairReader(reader io.Reader) *asciiCodePairReader {
	return &asciiCodePairReader{
		scanner: *bufio.NewScanner(reader),
	}
}

func (a *asciiCodePairReader) readLine() (line string, err error) {
	if !a.scanner.Scan() {
		err = a.scanner.Err()
		return
	}
	line = a.scanner.Text()
	return
}

func (a *asciiCodePairReader) readCode() (int, error) {
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

func (a *asciiCodePairReader) readCodePair() (CodePair, error) {
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
		codePair = NewStringCodePair(code, stringValue)
	}

	return codePair, nil
}
