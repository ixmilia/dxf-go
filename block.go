package dxf

type Block struct {
	handle         Handle
	endBlockHandle Handle
	IsInPaperSpace bool
	Layer          string
	Name           string
	BasePoint      Point
	XrefName       string
	Entities       []Entity
	Description    string
}

func NewBlock() *Block {
	return &Block{
		Layer: "0",
	}
}

func (b *Block) assignHandles(nextHandle uint32) uint32 {
	b.handle = Handle(nextHandle)
	nextHandle++
	b.endBlockHandle = Handle(nextHandle)
	nextHandle++

	for i := range b.Entities {
		e := &b.Entities[i]
		(*e).SetHandle(Handle(nextHandle))
		nextHandle++
	}

	return nextHandle
}

func (b *Block) getBlockPairs(version AcadVersion) (pairs []CodePair) {
	pairs = make([]CodePair, 0)
	pairs = append(pairs, NewStringCodePair(0, "BLOCK"))
	if version >= R13 {
		pairs = append(pairs, NewStringCodePair(5, stringFromHandle(b.handle)))
		pairs = append(pairs, NewStringCodePair(100, "AcDbEntity"))
	}
	if b.IsInPaperSpace {
		pairs = append(pairs, NewShortCodePair(67, shortFromBool(b.IsInPaperSpace)))
	}
	pairs = append(pairs, NewStringCodePair(8, b.Layer))
	if version >= R13 {
		pairs = append(pairs, NewStringCodePair(100, "AcDbBlockBegin"))
	}
	pairs = append(pairs, NewStringCodePair(2, b.Name))
	pairs = append(pairs, NewShortCodePair(70, 0)) // flags
	pairs = append(pairs, NewDoubleCodePair(10, b.BasePoint.X))
	pairs = append(pairs, NewDoubleCodePair(20, b.BasePoint.Y))
	pairs = append(pairs, NewDoubleCodePair(30, b.BasePoint.Z))
	if version >= R12 {
		pairs = append(pairs, NewStringCodePair(3, b.Name))
	}
	pairs = append(pairs, NewStringCodePair(1, b.XrefName))
	if len(b.Description) > 0 {
		pairs = append(pairs, NewStringCodePair(4, b.Description))
	}

	for i := range b.Entities {
		e := &b.Entities[i]
		pairs = append(pairs, allCodePairs(*e, version)...)
	}

	pairs = append(pairs, NewStringCodePair(0, "ENDBLK"))
	pairs = append(pairs, NewStringCodePair(5, stringFromHandle(b.endBlockHandle)))
	if version >= R13 {
		pairs = append(pairs, NewStringCodePair(100, "AcDbEntity"))
	}
	if b.IsInPaperSpace {
		pairs = append(pairs, NewShortCodePair(67, shortFromBool(b.IsInPaperSpace)))
	}
	pairs = append(pairs, NewStringCodePair(8, b.Layer))
	if version >= R13 {
		pairs = append(pairs, NewStringCodePair(100, "AcDbBlockEnd"))
	}

	return
}
