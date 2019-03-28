package dxf

// ViewMode represents the various states a given `ViewPort` can have.
type ViewMode int

// PerspectiveViewActive specifies whether the perspective view is active.
func (v *ViewMode) PerspectiveViewActive() bool {
	return int(*v)&1 != 0
}

// SetPerspectiveViewActive sets the active state of the perspective view.
func (v *ViewMode) SetPerspectiveViewActive(val bool) {
	if val {
		*v = ViewMode(int(*v) | 1)
	} else {
		*v = ViewMode(int(*v) & ^1)
	}
}

// FrontClippingOn specifies whether front clipping is on.
func (v *ViewMode) FrontClippingOn() bool {
	return int(*v)&2 != 0
}

// SetFrontClippingOn sets the front clipping state of the view.
func (v *ViewMode) SetFrontClippingOn(val bool) {
	if val {
		*v = ViewMode(int(*v) | 2)
	} else {
		*v = ViewMode(int(*v) & ^2)
	}
}

// BackClippingOn specifies whether back clipping is on.
func (v *ViewMode) BackClippingOn() bool {
	return int(*v)&4 != 0
}

// SetBackClippingOn sets the front clipping state of the view.
func (v *ViewMode) SetBackClippingOn(val bool) {
	if val {
		*v = ViewMode(int(*v) | 4)
	} else {
		*v = ViewMode(int(*v) & ^4)
	}
}

// UcsFollowModeOn specifies whether UCS follow mode is on.
func (v *ViewMode) UcsFollowModeOn() bool {
	return int(*v)&8 != 0
}

// SetUcsFollowModeOn sets the UCS follow mode of the view.
func (v *ViewMode) SetUcsFollowModeOn(val bool) {
	if val {
		*v = ViewMode(int(*v) | 8)
	} else {
		*v = ViewMode(int(*v) & ^8)
	}
}

// FrontClippingAtEye specifies whether front eye clipping is on.
func (v *ViewMode) FrontClippingAtEye() bool {
	return int(*v)&16 != 0
}

// SetFrontClippingAtEye sets the front eye clipping mode of the view.
func (v *ViewMode) SetFrontClippingAtEye(val bool) {
	if val {
		*v = ViewMode(int(*v) | 16)
	} else {
		*v = ViewMode(int(*v) & ^16)
	}
}
