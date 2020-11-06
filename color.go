package dxf

import "fmt"

// Color represents a color present in a DXF drawing.
type Color int16

// ByLayer creates a color that inherits from the item's layer color.
func ByLayer() Color {
	return Color(256)
}

// ByEntity creates a color that inherits from the item's entity.
func ByEntity() Color {
	return Color(257)
}

// ByBlock creates a color that inherits from the item's block.
func ByBlock() Color {
	return Color(0)
}

// ByLayer returns true if the color is inherited from the item's layer.
func (c *Color) ByLayer() bool {
	return int16(*c) == 256
}

// ByEntity returns true if the color is inherited from the item's entity.
func (c *Color) ByEntity() bool {
	return int16(*c) == 257
}

// ByBlock returns true if the color is inherited from the item's block.
func (c *Color) ByBlock() bool {
	return int16(*c) == 0
}

// TurnedOff returns true if the item's color is disabled.
func (c *Color) TurnedOff() bool {
	return int16(*c) < 0
}

// SetByLayer sets the color to inherit from the item's layer.
func (c *Color) SetByLayer() {
	*c = Color(256)
}

// SetByEntity sets the color to inherit from the item's entity.
func (c *Color) SetByEntity() {
	*c = Color(257)
}

// SetByBlock sets the color to inherit from the item's block.
func (c *Color) SetByBlock() {
	*c = Color(0)
}

// TurnOff sets the color to be disabled.
func (c *Color) TurnOff() {
	*c = Color(-1)
}

func (c *Color) String() string {
	switch {
	case c.ByLayer():
		return "BYLAYER"
	case c.ByEntity():
		return "BYENTITY"
	case c.ByBlock():
		return "BYBLOCK"
	case c.TurnedOff():
		return "OFF"
	default:
		return fmt.Sprint(int16(*c))
	}
}
