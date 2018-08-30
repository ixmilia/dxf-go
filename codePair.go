package dxf

type CodePair struct {
	Code  int
	Value CodePairValue
}

func NewStringCodePair(code int, value string) CodePair {
	return CodePair{
		Code:  code,
		Value: NewStringCodePairValue(value),
	}
}

type CodePairValue interface {
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
