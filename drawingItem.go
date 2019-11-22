package dxf

// DrawingItem represents an item in a DXF drawing that has a `Handle` value.
type DrawingItem interface {
	Handle() Handle
	SetHandle(h Handle)
	Owner() *DrawingItem
	SetOwner(val *DrawingItem)
}
