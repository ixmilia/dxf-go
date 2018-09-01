package dxf

type LineWeight int16

const (
	LineWeightStandard LineWeight = -3
	LineWeightByLayer  LineWeight = -2
	LineWeightByBlock  LineWeight = -1
)

func (l *LineWeight) SetStandard() {
	*l = LineWeightStandard
}

func (l *LineWeight) SetByLayer() {
	*l = LineWeightByLayer
}

func (l *LineWeight) SetByBlock() {
	*l = LineWeightByBlock
}

func (l *LineWeight) SetCustom(val int16) {
	*l = LineWeight(val)
}

func (l *LineWeight) Standard() bool {
	return *l == LineWeightStandard
}

func (l *LineWeight) ByLayer() bool {
	return *l == LineWeightByLayer
}

func (l *LineWeight) ByBlock() bool {
	return *l == LineWeightByBlock
}

func (l *LineWeight) Custom() bool {
	return int16(*l) >= 0
}

func NewLineWeightStandard() LineWeight {
	return LineWeight(int16(LineWeightStandard))
}

func NewLineWeightByLayer() LineWeight {
	return LineWeight(int16(LineWeightByLayer))
}

func NewLineWeightByBlock() LineWeight {
	return LineWeight(int16(LineWeightByBlock))
}
