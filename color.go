package dxf

type Color int16

func ByLayer() Color {
	return Color(256)
}

func ByEntity() Color {
	return Color(257)
}

func ByBlock() Color {
	return Color(0)
}

func (c *Color) ByLayer() bool {
	return int16(*c) == 256
}

func (c *Color) ByEntity() bool {
	return int16(*c) == 257
}

func (c *Color) ByBlock() bool {
	return int16(*c) == 0
}

func (c *Color) TurnedOff() bool {
	return int16(*c) < 0
}

func (c *Color) SetByLayer() {
	*c = Color(256)
}

func (c *Color) SetByEntity() {
	*c = Color(257)
}

func (c *Color) SetByBlock() {
	*c = Color(0)
}

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
		return string(int16(*c))
	}
}
