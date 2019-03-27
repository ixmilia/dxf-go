package dxf

import (
	"fmt"
)

// CodePair represents a code and value pair from a DXF drawing.
type CodePair struct {
	Code  int
	Value CodePairValue
}

func (pair *CodePair) String() string {
	return fmt.Sprintf("%d/%s", pair.Code, pair.Value)
}

func (pair *CodePair) isStartSection() bool {
	return pair.Code == 0 && pair.Value.(StringCodePairValue).Value == "SECTION"
}

func (pair *CodePair) isEndSection() bool {
	return pair.Code == 0 && pair.Value.(StringCodePairValue).Value == "ENDSEC"
}

func (pair *CodePair) isStartTable() bool {
	return pair.Code == 0 && pair.Value.(StringCodePairValue).Value == "TABLE"
}

func (pair *CodePair) isEndTable() bool {
	return pair.Code == 0 && pair.Value.(StringCodePairValue).Value == "ENDTAB"
}

func (pair *CodePair) isEOF() bool {
	return pair.Code == 0 && pair.Value.(StringCodePairValue).Value == "EOF"
}

// NewBoolCodePair creates a code pair representing a boolean value.
func NewBoolCodePair(code int, value bool) CodePair {
	return CodePair{
		Code:  code,
		Value: NewBoolCodePairValue(value),
	}
}

// NewDoubleCodePair creates a code pair representing a floating point value.
func NewDoubleCodePair(code int, value float64) CodePair {
	return CodePair{
		Code:  code,
		Value: NewDoubleCodePairValue(value),
	}
}

// NewIntCodePair creates a code pair representing an integer value.
func NewIntCodePair(code int, value int) CodePair {
	return CodePair{
		Code:  code,
		Value: NewIntCodePairValue(value),
	}
}

// NewLongCodePair creates a code pair representing a long integer value.
func NewLongCodePair(code int, value int64) CodePair {
	return CodePair{
		Code:  code,
		Value: NewLongCodePairValue(value),
	}
}

// NewShortCodePair creates a code pair representing a short integer value.
func NewShortCodePair(code int, value int16) CodePair {
	return CodePair{
		Code:  code,
		Value: NewShortCodePairValue(value),
	}
}

// NewStringCodePair creates a code pair representing a string value.
func NewStringCodePair(code int, value string) CodePair {
	return CodePair{
		Code:  code,
		Value: NewStringCodePairValue(value),
	}
}

// CodePairValue represents a value in a single code pair.
type CodePairValue interface {
}

// BoolCodePairValue represents a boolean code pair value.
type BoolCodePairValue struct {
	CodePairValue
	Value bool
}

// NewBoolCodePairValue creates a boolean code pair value.
func NewBoolCodePairValue(value bool) CodePairValue {
	return BoolCodePairValue{
		Value: value,
	}
}

// DoubleCodePairValue represents a floating point code pair value.
type DoubleCodePairValue struct {
	CodePairValue
	Value float64
}

// NewDoubleCodePairValue creates a floating point code pair value.
func NewDoubleCodePairValue(value float64) CodePairValue {
	return DoubleCodePairValue{
		Value: value,
	}
}

// IntCodePairValue represents an integer code pair value.
type IntCodePairValue struct {
	CodePairValue
	Value int
}

// NewIntCodePairValue creates an integer code pair value.
func NewIntCodePairValue(value int) CodePairValue {
	return IntCodePairValue{
		Value: value,
	}
}

// LongCodePairValue represents a long integer code pair value.
type LongCodePairValue struct {
	CodePairValue
	Value int64
}

// NewLongCodePairValue creates a long integer code pair value.
func NewLongCodePairValue(value int64) CodePairValue {
	return LongCodePairValue{
		Value: value,
	}
}

// ShortCodePairValue represents a short integer code pair value.
type ShortCodePairValue struct {
	CodePairValue
	Value int16
}

// NewShortCodePairValue creates a short integer code pair value.
func NewShortCodePairValue(value int16) CodePairValue {
	return ShortCodePairValue{
		Value: value,
	}
}

// StringCodePairValue represents a string code pair value.
type StringCodePairValue struct {
	CodePairValue
	Value string
}

// NewStringCodePairValue creates a string code pair value.
func NewStringCodePairValue(value string) CodePairValue {
	return StringCodePairValue{
		Value: value,
	}
}
