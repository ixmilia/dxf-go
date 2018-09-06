package dxf

// LineWeight represents a plotted drawing elements line weight.
type LineWeight int16

const (
	// LineWeightStandard represents the standard line weight.
	LineWeightStandard LineWeight = -3

	// LineWeightByLayer represents a line weight inherited from the item's layer.
	LineWeightByLayer LineWeight = -2

	// LineWeightByBlock represents a line weight inherited from the item's block.
	LineWeightByBlock LineWeight = -1
)

// SetStandard sets the line weight to the standard value.
func (l *LineWeight) SetStandard() {
	*l = LineWeightStandard
}

// SetByLayer sets the line weight to inherit from the item's layer.
func (l *LineWeight) SetByLayer() {
	*l = LineWeightByLayer
}

// SetByBlock sets the line weight to inherit from the item's block.
func (l *LineWeight) SetByBlock() {
	*l = LineWeightByBlock
}

// SetCustom sets the line weight to a custom value.
func (l *LineWeight) SetCustom(val int16) {
	*l = LineWeight(val)
}

// Standard returns true if the line weight is the standard value.
func (l *LineWeight) Standard() bool {
	return *l == LineWeightStandard
}

// ByLayer returns true if the line weight is inherited from the item's layer.
func (l *LineWeight) ByLayer() bool {
	return *l == LineWeightByLayer
}

// ByBlock returns true if the line weight is inherited from the item's block.
func (l *LineWeight) ByBlock() bool {
	return *l == LineWeightByBlock
}

// Custom returns true if the line weight is a custom value.  To retreive the custom value, use `int16(lineWeight)`.
func (l *LineWeight) Custom() bool {
	return int16(*l) >= 0
}

// NewLineWeightStandard creates a new standard line weight.
func NewLineWeightStandard() LineWeight {
	return LineWeight(int16(LineWeightStandard))
}

// NewLineWeightByLayer creates a line weight that is inherited from the item's layer.
func NewLineWeightByLayer() LineWeight {
	return LineWeight(int16(LineWeightByLayer))
}

// NewLineWeightByBlock creates a line weight that is inherited from the item's block.
func NewLineWeightByBlock() LineWeight {
	return LineWeight(int16(LineWeightByBlock))
}
