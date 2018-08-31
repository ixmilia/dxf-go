package dxf

import (
	"fmt"
)

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

func (pair *CodePair) isEof() bool {
	return pair.Code == 0 && pair.Value.(StringCodePairValue).Value == "EOF"
}

func NewBoolCodePair(code int, value bool) CodePair {
	return CodePair{
		Code:  code,
		Value: NewBoolCodePairValue(value),
	}
}

func NewDoubleCodePair(code int, value float64) CodePair {
	return CodePair{
		Code:  code,
		Value: NewDoubleCodePairValue(value),
	}
}

func NewIntCodePair(code int, value int) CodePair {
	return CodePair{
		Code:  code,
		Value: NewIntCodePairValue(value),
	}
}

func NewLongCodePair(code int, value int64) CodePair {
	return CodePair{
		Code:  code,
		Value: NewLongCodePairValue(value),
	}
}

func NewShortCodePair(code int, value int16) CodePair {
	return CodePair{
		Code:  code,
		Value: NewShortCodePairValue(value),
	}
}

func NewStringCodePair(code int, value string) CodePair {
	return CodePair{
		Code:  code,
		Value: NewStringCodePairValue(value),
	}
}

type CodePairValue interface {
}

type BoolCodePairValue struct {
	CodePairValue
	Value bool
}

func NewBoolCodePairValue(value bool) CodePairValue {
	return BoolCodePairValue{
		Value: value,
	}
}

type DoubleCodePairValue struct {
	CodePairValue
	Value float64
}

func NewDoubleCodePairValue(value float64) CodePairValue {
	return DoubleCodePairValue{
		Value: value,
	}
}

type IntCodePairValue struct {
	CodePairValue
	Value int
}

func NewIntCodePairValue(value int) CodePairValue {
	return IntCodePairValue{
		Value: value,
	}
}

type LongCodePairValue struct {
	CodePairValue
	Value int64
}

func NewLongCodePairValue(value int64) CodePairValue {
	return LongCodePairValue{
		Value: value,
	}
}

type ShortCodePairValue struct {
	CodePairValue
	Value int16
}

func NewShortCodePairValue(value int16) CodePairValue {
	return ShortCodePairValue{
		Value: value,
	}
}

type StringCodePairValue struct {
	CodePairValue
	Value string
}

func NewStringCodePairValue(value string) CodePairValue {
	return StringCodePairValue{
		Value: value,
	}
}
