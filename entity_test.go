package dxf

import (
	"testing"
)

func TestParseSimpleLine(t *testing.T) {
	line := parseEntity(t, "LINE",
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(11, 4.0),
		NewDoubleCodePair(21, 5.0),
		NewDoubleCodePair(31, 6.0),
	).(*Line)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, line.P1)
	assertEqPoint(t, Point{4.0, 5.0, 6.0}, line.P2)
}

func TestParseAlternateTypeString(t *testing.T) {
	line := parseEntity(t, "3DLINE",
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(11, 4.0),
		NewDoubleCodePair(21, 5.0),
		NewDoubleCodePair(31, 6.0),
	).(*Line)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, line.P1)
	assertEqPoint(t, Point{4.0, 5.0, 6.0}, line.P2)
}

func TestParseUnsupportedEntity(t *testing.T) {
	drawing := parseFromCodePairs(t,
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "ENTITIES"),
		NewStringCodePair(0, "LINE"),          // supported entity
		NewStringCodePair(0, "NOT_AN_ENTITY"), // unsupported entity
		NewStringCodePair(0, "LINE"),          // supported entity
		NewStringCodePair(0, "ENDSEC"),
		NewStringCodePair(0, "EOF"),
	)
	assertEqInt(t, 2, len(drawing.Entities))
	assertEqString(t, "LINE", drawing.Entities[0].typeString())
	assertEqString(t, "LINE", drawing.Entities[1].typeString())
}

func TestWriteSimpleLine(t *testing.T) {
	/*
		testHelpers.go:55: Unable to find '
		10/{<nil> %!s(float64=1)}
		20/{<nil> %!s(float64=2)}
		30/{<nil> %!s(float64=3)}
		11/{<nil> %!s(float64=4)}
		21/{<nil> %!s(float64=5)}
		31/{<nil> %!s(float64=6)}' in '
		0/{<nil> LINE}
		100/{<nil> AcDbEntity}
		8/{<nil> 0}
		100/{<nil> AcDbLine}
		10/{<nil> %!s(float64=1)}
		20/{<nil> %!s(float64=2)}
		30/{<nil> %!s(float64=3)}
		11/{<nil> %!s(float64=4)}
		21/{<nil> %!s(float64=5)}
		31/{<nil> %!s(float64=6)}'
	*/
	line := NewLine()
	line.P1 = Point{1.0, 2.0, 3.0}
	line.P2 = Point{4.0, 5.0, 6.0}
	actual := allCodePairs(line, R12)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "LINE"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(11, 4.0),
		NewDoubleCodePair(21, 5.0),
		NewDoubleCodePair(31, 6.0),
	}, actual)
}

func TestConditionalEntityFieldWriting(t *testing.T) {
	line := NewLine()
	line.SetIsInPaperSpace(false)
	actual := allCodePairs(line, R14)
	assertNotContainsCodePairs(t, []CodePair{
		NewShortCodePair(67, 0), // this is only written when Version >= R12 and it's not the default (false)
	}, actual)
}

func TestReadEntityFieldFlag(t *testing.T) {
	face := parseEntity(t, "3DFACE",
		NewShortCodePair(70, 5),
	).(*Face)
	assert(t, face.FirstEdgeInvisible(), "expected first edge to be invisible")
	assert(t, !face.SecondEdgeInvisible(), "expected first edge to be visible")
	assert(t, face.ThirdEdgeInvisible(), "expected first edge to be invisible")
	assert(t, !face.FourthEdgeInvisible(), "expected first edge to be visible")
}

func TestWriteEntityFieldFlag(t *testing.T) {
	face := NewFace()
	face.SetFirstEdgeInvisible(true)
	face.SetSecondEdgeInvisible(false)
	face.SetThirdEdgeInvisible(true)
	face.SetFourthEdgeInvisible(false)
	actual := allCodePairs(face, R12)
	assertContainsCodePairs(t, []CodePair{
		NewShortCodePair(70, 0b0101),
	}, actual)
}

func TestWriteVersionSpecificEntities(t *testing.T) {
	solid := NewSolid3D()
	drawing := *NewDrawing()
	drawing.Entities = append(drawing.Entities, solid)

	// ensure it's present when appropriate
	drawing.Header.Version = R13
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "3DSOLID"),
	}, drawingCodePairs(t, drawing))

	// and not otherwise
	drawing.Header.Version = R12
	assertNotContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "3DSOLID"),
	}, drawingCodePairs(t, drawing))
}

func TestReadMultipleBaseEntityData(t *testing.T) {
	line := parseEntity(t, "LINE",
		NewStringCodePair(310, "line 1"),
		NewStringCodePair(310, "line 2"),
	).(*Line)
	assertEqInt(t, 2, len(line.PreviewImageData()))
	assertEqString(t, "line 1", line.PreviewImageData()[0])
	assertEqString(t, "line 2", line.PreviewImageData()[1])
}

func TestWriteMultipleBaseEntityData(t *testing.T) {
	line := NewLine()
	line.SetPreviewImageData(append(line.PreviewImageData(), "line 1"))
	line.SetPreviewImageData(append(line.PreviewImageData(), "line 2"))
	actual := allCodePairs(line, R2000)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(310, "line 1"),
		NewStringCodePair(310, "line 2"),
		NewStringCodePair(100, "AcDbLine"),
	}, actual)
}

func TestReadMultipleSpecificEntityData(t *testing.T) {
	solid := parseEntity(t, "3DSOLID",
		NewStringCodePair(1, "line 1"),
		NewStringCodePair(1, "line 2"),
	).(*Solid3D)
	assertEqInt(t, 2, len(solid.CustomData))
	assertEqString(t, "line 1", solid.CustomData[0])
	assertEqString(t, "line 2", solid.CustomData[1])
}

func TestWriteMultipleSpecificEntityData(t *testing.T) {
	solid := NewSolid3D()
	solid.AddCustomData("line 1")
	solid.AddCustomData("line 2")
	actual := allCodePairs(solid, R13)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDbModelerGeometry"),
		NewShortCodePair(70, 1),
		NewStringCodePair(1, "line 1"),
		NewStringCodePair(1, "line 2"),
	}, actual)
}

func TestWriteConditionsOnWriteOrderDirectives(t *testing.T) {
	solid := NewSolid3D()
	solid.AddCustomData("custom data")

	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDb3dSolid"),
	}, drawingCodePairsFromEntity(t, solid, R2007))

	assertNotContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDb3dSolid"),
	}, drawingCodePairsFromEntity(t, solid, R13))
}

func TestReadEntityWithCustomReader(t *testing.T) {
	proxy := parseEntity(t, "ACAD_PROXY_ENTITY",
		NewIntCodePair(92, 4),
		NewStringCodePair(310, "1234"),
		NewStringCodePair(310, "ABCD"),
		NewIntCodePair(93, 4),
		NewStringCodePair(310, "5678"),
		NewStringCodePair(310, "DCBA"),
	).(*ProxyEntity)
	assertEqByteArray(t, []byte{0x12, 0x34, 0xAB, 0xCD}, proxy.GraphicsData)
	assertEqByteArray(t, []byte{0x56, 0x78, 0xDC, 0xBA}, proxy.EntityData)
}

func TestWriteEntityWithBeforeWrite(t *testing.T) {
	proxy := NewProxyEntity()
	proxy.GraphicsData = []byte{0x12, 0x34, 0xAB, 0xCD}
	proxy.EntityData = []byte{0x56, 0x78, 0xDC, 0xBA}
	actual := allCodePairs(proxy, R14)
	assertContainsCodePairs(t, []CodePair{
		NewIntCodePair(92, 4),
		NewStringCodePair(310, "1234ABCD"),
		NewIntCodePair(93, 4),
		NewStringCodePair(310, "5678DCBA"),
	}, actual)
}

func TestReadCollectedEntities(t *testing.T) {
	attdef := parseEntity(t, "ATTDEF",
		NewStringCodePair(0, "MTEXT"),
		NewStringCodePair(1, "mtext-value"),
	).(*AttributeDefinition)
	assertEqString(t, "mtext-value", attdef.MText.Text)
}

func TestWriteEntityWithTrailingEntities(t *testing.T) {
	attdef := NewAttributeDefinition()
	actual := drawingCodePairsFromEntity(t, attdef, R14)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "ATTDEF"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "MTEXT"),
	}, actual)
}

func TestReadDimension(t *testing.T) {
	dim := parseEntity(t, "DIMENSION",
		NewStringCodePair(1, "text"),
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewShortCodePair(70, 1), // aligned
		NewStringCodePair(100, "AcDbAlignedDimension"),
		NewDoubleCodePair(13, 3.0),
		NewDoubleCodePair(23, 4.0),
		NewDoubleCodePair(14, 5.0),
		NewDoubleCodePair(24, 6.0),
	).(*AlignedDimension)
	assertEqString(t, "text", dim.Text())
	assertEqPoint(t, Point{1.0, 2.0, 0.0}, dim.DefinitionPoint1())
	assertEqPoint(t, Point{3.0, 4.0, 0.0}, dim.DefinitionPoint2)
	assertEqPoint(t, Point{5.0, 6.0, 0.0}, dim.DefinitionPoint3)
}

func TestWriteDimension(t *testing.T) {
	dim := NewAlignedDimension()
	dim.SetDefinitionPoint1(Point{1.0, 2.0, 0.0})
	dim.DefinitionPoint2 = Point{3.0, 4.0, 0.0}
	dim.DefinitionPoint3 = Point{5.0, 6.0, 0.0}
	actual := allCodePairs(dim, R14)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 0.0),
		NewDoubleCodePair(11, 0.0),
		NewDoubleCodePair(21, 0.0),
		NewDoubleCodePair(31, 0.0),
		NewShortCodePair(70, 1),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDbAlignedDimension"),
		NewDoubleCodePair(13, 3.0),
		NewDoubleCodePair(23, 4.0),
		NewDoubleCodePair(33, 0.0),
		NewDoubleCodePair(14, 5.0),
		NewDoubleCodePair(24, 6.0),
		NewDoubleCodePair(34, 0.0),
	}, actual)
}

func TestReadImage(t *testing.T) {
	img := parseEntity(t, "IMAGE",
		NewIntCodePair(91, 2),
		NewDoubleCodePair(14, 1.0),
		NewDoubleCodePair(24, 2.0),
		NewDoubleCodePair(14, 3.0),
		NewDoubleCodePair(24, 4.0),
	).(*Image)
	assertEqInt(t, 2, len(img.ClippingVertices()))
	assertEqPoint(t, Point{1.0, 2.0, 0.0}, img.ClippingVertices()[0])
	assertEqPoint(t, Point{3.0, 4.0, 0.0}, img.ClippingVertices()[1])
}

func TestWriteImage(t *testing.T) {
	img := NewImage()
	img.SetClippingVertices(append(img.ClippingVertices(), Point{1.0, 2.0, 0.0}))
	img.SetClippingVertices(append(img.ClippingVertices(), Point{3.0, 4.0, 0.0}))
	actual := allCodePairs(img, R14)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDbRasterImage"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewIntCodePair(91, 2),
		NewDoubleCodePair(14, 1.0),
		NewDoubleCodePair(24, 2.0),
		NewDoubleCodePair(14, 3.0),
		NewDoubleCodePair(24, 4.0),
	}, actual)
}

func TestReadInsertAtEnd(t *testing.T) {
	ins := parseEntity(t, "INSERT",
		NewShortCodePair(66, 1), // has attributes
		NewStringCodePair(0, "ATTRIB"),
		NewStringCodePair(1, "attrib 1"),
		NewStringCodePair(0, "ATTRIB"),
		NewStringCodePair(1, "attrib 2"),
		NewStringCodePair(0, "SEQEND"),
	).(*Insert)
	assertEqInt(t, 2, len(ins.Attributes))
	assertEqString(t, "attrib 1", ins.Attributes[0].Value)
	assertEqString(t, "attrib 2", ins.Attributes[1].Value)
}

func TestReadInsertAtEndNoSeqend(t *testing.T) {
	ins := parseEntity(t, "INSERT",
		NewShortCodePair(66, 1), // has attributes
		NewStringCodePair(0, "ATTRIB"),
		NewStringCodePair(1, "attrib 1"),
		NewStringCodePair(0, "ATTRIB"),
		NewStringCodePair(1, "attrib 2"),
	).(*Insert)
	assertEqInt(t, 2, len(ins.Attributes))
	assertEqString(t, "attrib 1", ins.Attributes[0].Value)
	assertEqString(t, "attrib 2", ins.Attributes[1].Value)
}

func TestReadInsertWithTrailingEntity(t *testing.T) {
	entities := parseEntities(t,
		NewStringCodePair(0, "INSERT"),
		NewShortCodePair(66, 1), // has attributes
		NewStringCodePair(0, "ATTRIB"),
		NewStringCodePair(1, "attrib 1"),
		NewStringCodePair(0, "ATTRIB"),
		NewStringCodePair(1, "attrib 2"),
		NewStringCodePair(0, "SEQEND"),
		NewStringCodePair(0, "LINE"), // trailing entity
		NewDoubleCodePair(10, 11.0),
	)
	assertEqInt(t, 2, len(entities))
	ins := entities[0].(*Insert)
	assertEqInt(t, 2, len(ins.Attributes))
	assertEqString(t, "attrib 1", ins.Attributes[0].Value)
	assertEqString(t, "attrib 2", ins.Attributes[1].Value)
	line := entities[1].(*Line)
	assertEqPoint(t, Point{11.0, 0.0, 0.0}, line.P1)
}

func TestReadInsertWithTrailingEntityNoSeqend(t *testing.T) {
	entities := parseEntities(t,
		NewStringCodePair(0, "INSERT"),
		NewShortCodePair(66, 1), // has attributes
		NewStringCodePair(0, "ATTRIB"),
		NewStringCodePair(1, "attrib 1"),
		NewStringCodePair(0, "ATTRIB"),
		NewStringCodePair(1, "attrib 2"),
		NewStringCodePair(0, "LINE"), // trailing entity
		NewDoubleCodePair(10, 11.0),
	)
	assertEqInt(t, 2, len(entities))
	ins := entities[0].(*Insert)
	assertEqInt(t, 2, len(ins.Attributes))
	assertEqString(t, "attrib 1", ins.Attributes[0].Value)
	assertEqString(t, "attrib 2", ins.Attributes[1].Value)
	line := entities[1].(*Line)
	assertEqPoint(t, Point{11.0, 0.0, 0.0}, line.P1)
}

func TestWriteInsert(t *testing.T) {
	ins := NewInsert()
	att1 := *NewAttribute()
	att1.Value = "attrib 1"
	ins.Attributes = append(ins.Attributes, att1)
	att2 := *NewAttribute()
	att2.Value = "attrib 2"
	ins.Attributes = append(ins.Attributes, att2)
	actual := allCodePairs(ins, R14)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(1, "attrib 1"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(1, "attrib 2"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "SEQEND"),
	}, actual)
}

func TestReadLWPolyline(t *testing.T) {
	lw := parseEntity(t, "LWPOLYLINE",
		NewShortCodePair(70, 1),
		NewIntCodePair(90, 2),      // 2 vertices
		NewDoubleCodePair(10, 1.0), // v1
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(10, 3.0), // v2
		NewDoubleCodePair(20, 4.0),
		NewIntCodePair(91, 42),
	).(*LWPolyline)
	assert(t, lw.IsClosed(), "expected LWPOLYLINE to be closed")
	assertEqInt(t, 2, len(lw.Vertices))
	assertEqFloat64(t, 1.0, lw.Vertices[0].X)
	assertEqFloat64(t, 2.0, lw.Vertices[0].Y)
	assertEqInt(t, 0, lw.Vertices[0].ID)
	assertEqFloat64(t, 3.0, lw.Vertices[1].X)
	assertEqFloat64(t, 4.0, lw.Vertices[1].Y)
	assertEqInt(t, 42, lw.Vertices[1].ID)
}

func TestWriteLWPolyline(t *testing.T) {
	lw := NewLWPolyline()
	lw.Vertices = append(lw.Vertices, LwVertex{X: 1.0, Y: 2.0})
	lw.Vertices = append(lw.Vertices, LwVertex{X: 3.0, Y: 4.0, ID: 42})
	actual := allCodePairs(lw, R2013)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(10, 1.0), // v1
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(10, 3.0), // v2
		NewDoubleCodePair(20, 4.0),
		NewIntCodePair(91, 42),
	}, actual)
}

func TestReadModelPoint(t *testing.T) {
	p := parseEntity(t, "POINT",
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
	).(*ModelPoint)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, p.Location)
}

func TestWriteModelPoint(t *testing.T) {
	p := NewModelPoint()
	p.Location = Point{1.0, 2.0, 3.0}
	actual := allCodePairs(p, R14)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDbPoint"),
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
	}, actual)
}

func TestReadPolylineWithNoVertices(t *testing.T) {
	p := parseEntity(t, "POLYLINE",
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewStringCodePair(0, "SEQEND"),
	).(*Polyline)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, p.Location)
	assertEqInt(t, 0, len(p.Vertices))
}

func TestReadPolylineWithCLOValues(t *testing.T) {
	p := parseEntity(t, "POLYLINE",
		NewShortCodePair(250, 2),
	).(*Polyline)
	assertEqShort(t, 2, int16(p.CLO_PolylineType))
}

func TestReadPolylineWithMutlipleVertices(t *testing.T) {
	p := parseEntity(t, "POLYLINE",
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewStringCodePair(0, "VERTEX"),
		NewDoubleCodePair(10, 11.0),
		NewDoubleCodePair(20, 22.0),
		NewDoubleCodePair(30, 33.0),
		NewStringCodePair(0, "VERTEX"),
		NewDoubleCodePair(10, 111.0),
		NewDoubleCodePair(20, 222.0),
		NewDoubleCodePair(30, 333.0),
		NewStringCodePair(0, "SEQEND"),
	).(*Polyline)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, p.Location)
	assertEqInt(t, 2, len(p.Vertices))
	assertEqPoint(t, Point{11.0, 22.0, 33.0}, p.Vertices[0].Location)
	assertEqPoint(t, Point{111.0, 222.0, 333.0}, p.Vertices[1].Location)
}

func TestReadPolylineWithNoVerticesNoSeqend(t *testing.T) {
	p := parseEntity(t, "POLYLINE",
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
	).(*Polyline)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, p.Location)
	assertEqInt(t, 0, len(p.Vertices))
}

func TestReadPolylineWithMultipleVerticesNoSeqend(t *testing.T) {
	p := parseEntity(t, "POLYLINE",
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewStringCodePair(0, "VERTEX"),
		NewDoubleCodePair(10, 11.0),
		NewDoubleCodePair(20, 22.0),
		NewDoubleCodePair(30, 33.0),
		NewStringCodePair(0, "VERTEX"),
		NewDoubleCodePair(10, 111.0),
		NewDoubleCodePair(20, 222.0),
		NewDoubleCodePair(30, 333.0),
	).(*Polyline)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, p.Location)
	assertEqInt(t, 2, len(p.Vertices))
	assertEqPoint(t, Point{11.0, 22.0, 33.0}, p.Vertices[0].Location)
	assertEqPoint(t, Point{111.0, 222.0, 333.0}, p.Vertices[1].Location)
}

func TestReadPolylineWithNoVerticesNoSeqendTrailingEntity(t *testing.T) {
	entities := parseEntities(t,
		NewStringCodePair(0, "POLYLINE"),
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewStringCodePair(0, "LINE"),
		NewDoubleCodePair(10, 11.0),
		NewDoubleCodePair(20, 22.0),
		NewDoubleCodePair(30, 33.0),
	)
	assertEqInt(t, 2, len(entities))

	p := entities[0].(*Polyline)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, p.Location)
	assertEqInt(t, 0, len(p.Vertices))

	l := entities[1].(*Line)
	assertEqPoint(t, Point{11.0, 22.0, 33.0}, l.P1)
}

func TestReadPolylineWithMultipleVerticesNoSeqendTrailingEntity(t *testing.T) {
	entities := parseEntities(t,
		NewStringCodePair(0, "POLYLINE"),
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewStringCodePair(0, "VERTEX"),
		NewDoubleCodePair(10, 11.0),
		NewDoubleCodePair(20, 22.0),
		NewDoubleCodePair(30, 33.0),
		NewStringCodePair(0, "VERTEX"),
		NewDoubleCodePair(10, 111.0),
		NewDoubleCodePair(20, 222.0),
		NewDoubleCodePair(30, 333.0),
		NewStringCodePair(0, "LINE"),
		NewDoubleCodePair(10, 11.0),
		NewDoubleCodePair(20, 22.0),
		NewDoubleCodePair(30, 33.0),
	)
	assertEqInt(t, 2, len(entities))

	p := entities[0].(*Polyline)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, p.Location)
	assertEqInt(t, 2, len(p.Vertices))
	assertEqPoint(t, Point{11.0, 22.0, 33.0}, p.Vertices[0].Location)
	assertEqPoint(t, Point{111.0, 222.0, 333.0}, p.Vertices[1].Location)

	l := entities[1].(*Line)
	assertEqPoint(t, Point{11.0, 22.0, 33.0}, l.P1)
}

func TestWrite2DPolylineTest(t *testing.T) {
	p := NewPolyline()
	p.Vertices = append(p.Vertices, *NewVertex())
	actual := allCodePairs(p, R14)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDb2dPolyline"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDbVertex"),
		NewStringCodePair(100, "AcDb2dVertex"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "SEQEND"),
	}, actual)
}

func TestWrite3DPolylineTest(t *testing.T) {
	p := NewPolyline()
	v := *NewVertex()
	v.Location.X = 1.0
	v.Location.Y = 2.0
	v.Location.Z = 3.0
	p.Vertices = append(p.Vertices, v)
	p.SetIs3DPolyline(true)
	actual := allCodePairs(p, R14)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDb3dPolyline"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDbVertex"),
		NewStringCodePair(100, "AcDb3dPolylineVertex"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "SEQEND"),
	}, actual)
}

func TestRoundTripPolylineTest(t *testing.T) {
	p := NewPolyline()
	v := *NewVertex()
	v.Location.X = 1.0
	v.Location.Y = 2.0
	v.Location.Z = 3.0
	p.Vertices = append(p.Vertices, v)
	actual := drawingCodePairsFromEntity(t, p, R14)

	drawing := parseFromCodePairs(t, actual...)
	assertEqInt(t, 1, len(drawing.Entities))
	p2 := drawing.Entities[0].(*Polyline)
	assertEqInt(t, 1, len(p2.Vertices))
	assertEqPoint(t, v.Location, p2.Vertices[0].Location)
}

func TestReadSection(t *testing.T) {
	s := parseEntity(t, "SECTION",
		// 3 vertices
		NewIntCodePair(92, 3),
		NewDoubleCodePair(11, 1.0),
		NewDoubleCodePair(21, 2.0),
		NewDoubleCodePair(31, 3.0),
		NewDoubleCodePair(11, 11.0),
		NewDoubleCodePair(21, 22.0),
		NewDoubleCodePair(31, 33.0),
		NewDoubleCodePair(11, 111.0),
		NewDoubleCodePair(21, 222.0),
		NewDoubleCodePair(31, 333.0),
		// 1 back vertex
		NewIntCodePair(93, 1),
		NewDoubleCodePair(12, 4.0),
		NewDoubleCodePair(22, 5.0),
		NewDoubleCodePair(32, 6.0),
	).(*Section)
	assertEqInt(t, 3, len(s.Vertices))
	assertEqPoint(t, s.Vertices[0], Point{X: 1.0, Y: 2.0, Z: 3.0})
	assertEqPoint(t, s.Vertices[1], Point{X: 11.0, Y: 22.0, Z: 33.0})
	assertEqPoint(t, s.Vertices[2], Point{X: 111.0, Y: 222.0, Z: 333.0})
	assertEqInt(t, 1, len(s.BackLineVertices))
	assertEqPoint(t, s.BackLineVertices[0], Point{X: 4.0, Y: 5.0, Z: 6.0})
}

func TestWriteSection(t *testing.T) {
	s := NewSection()
	s.Vertices = append(s.Vertices, Point{X: 1.0, Y: 2.0, Z: 3.0})
	s.Vertices = append(s.Vertices, Point{X: 11.0, Y: 22.0, Z: 33.0})
	s.Vertices = append(s.Vertices, Point{X: 111.0, Y: 222.0, Z: 333.0})
	s.BackLineVertices = append(s.BackLineVertices, Point{X: 4.0, Y: 5.0, Z: 6.0})
	actual := allCodePairs(s, R2007)
	assertContainsCodePairs(t, []CodePair{
		// 3 vertices
		NewIntCodePair(92, 3),
		NewDoubleCodePair(11, 1.0),
		NewDoubleCodePair(21, 2.0),
		NewDoubleCodePair(31, 3.0),
		NewDoubleCodePair(11, 11.0),
		NewDoubleCodePair(21, 22.0),
		NewDoubleCodePair(31, 33.0),
		NewDoubleCodePair(11, 111.0),
		NewDoubleCodePair(21, 222.0),
		NewDoubleCodePair(31, 333.0),
		// 1 back vertex
		NewIntCodePair(93, 1),
		NewDoubleCodePair(12, 4.0),
		NewDoubleCodePair(22, 5.0),
		NewDoubleCodePair(32, 6.0),
	}, actual)
}

func TestReadSplineWithWeights(t *testing.T) {
	s := parseEntity(t, "SPLINE",
		NewShortCodePair(73, 2),
		NewDoubleCodePair(41, 7.0),
		NewDoubleCodePair(41, 8.0),
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(10, 4.0),
		NewDoubleCodePair(20, 5.0),
		NewDoubleCodePair(30, 6.0),
	).(*Spline)
	assertEqInt(t, 2, len(s.ControlPoints))
	assertEqFloat64(t, 7.0, s.ControlPoints[0].Weight)
	assertEqPoint(t, Point{X: 1.0, Y: 2.0, Z: 3.0}, s.ControlPoints[0].Point)
	assertEqFloat64(t, 8.0, s.ControlPoints[1].Weight)
	assertEqPoint(t, Point{X: 4.0, Y: 5.0, Z: 6.0}, s.ControlPoints[1].Point)
}

func TestReadSplineWithoutWeights(t *testing.T) {
	s := parseEntity(t, "SPLINE",
		NewShortCodePair(73, 2),
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(10, 4.0),
		NewDoubleCodePair(20, 5.0),
		NewDoubleCodePair(30, 6.0),
	).(*Spline)
	assertEqInt(t, 2, len(s.ControlPoints))
	assertEqFloat64(t, 1.0, s.ControlPoints[0].Weight)
	assertEqPoint(t, Point{X: 1.0, Y: 2.0, Z: 3.0}, s.ControlPoints[0].Point)
	assertEqFloat64(t, 1.0, s.ControlPoints[1].Weight)
	assertEqPoint(t, Point{X: 4.0, Y: 5.0, Z: 6.0}, s.ControlPoints[1].Point)
}

func TestWriteSplineWithStandardWeights(t *testing.T) {
	s := NewSpline()
	s.ControlPoints = append(s.ControlPoints, ControlPoint{Point: Point{X: 1.0, Y: 2.0, Z: 3.0}, Weight: 1.0})
	s.ControlPoints = append(s.ControlPoints, ControlPoint{Point: Point{X: 4.0, Y: 5.0, Z: 6.0}, Weight: 1.0})
	actual := allCodePairs(s, R13)
	assertContainsCodePairs(t, []CodePair{
		NewShortCodePair(73, 2),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(10, 4.0),
		NewDoubleCodePair(20, 5.0),
		NewDoubleCodePair(30, 6.0),
	}, actual)
	assertNotContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(41, 1.0),
		NewDoubleCodePair(41, 1.0),
	}, actual)
}

func TestWriteSplineWithNonStandardWeights(t *testing.T) {
	s := NewSpline()
	s.ControlPoints = append(s.ControlPoints, ControlPoint{Point: Point{X: 1.0, Y: 2.0, Z: 3.0}, Weight: 7.0})
	s.ControlPoints = append(s.ControlPoints, ControlPoint{Point: Point{X: 4.0, Y: 5.0, Z: 6.0}, Weight: 8.0})
	actual := allCodePairs(s, R13)
	assertContainsCodePairs(t, []CodePair{
		NewShortCodePair(73, 2),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(10, 4.0),
		NewDoubleCodePair(20, 5.0),
		NewDoubleCodePair(30, 6.0),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(41, 7.0),
		NewDoubleCodePair(41, 8.0),
	}, actual)
}

func TestReadUnderlay(t *testing.T) {
	u := parseEntity(t, "DGNUNDERLAY",
		NewDoubleCodePair(10, 1.0), // insertion point
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(11, 4.0), // boundary points
		NewDoubleCodePair(21, 5.0),
		NewDoubleCodePair(11, 6.0),
		NewDoubleCodePair(21, 7.0),
	).(*DgnUnderlay)
	assertEqPoint(t, Point{X: 1.0, Y: 2.0, Z: 3.0}, u.InsertionPoint())
	assertEqInt(t, 2, len(u.BoundaryPoints()))
	assertEqPoint(t, Point{X: 4.0, Y: 5.0, Z: 0.0}, u.BoundaryPoints()[0])
	assertEqPoint(t, Point{X: 6.0, Y: 7.0, Z: 0.0}, u.BoundaryPoints()[1])
}

func TestWriteUnderlay(t *testing.T) {
	u := NewDgnUnderlay()
	u.SetInsertionPoint(Point{X: 1.0, Y: 2.0, Z: 3.0})
	u.SetBoundaryPoints(append(u.BoundaryPoints(), Point{X: 4.0, Y: 5.0, Z: 0.0}))
	u.SetBoundaryPoints(append(u.BoundaryPoints(), Point{X: 6.0, Y: 7.0, Z: 0.0}))
	actual := allCodePairs(u, R14)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(0, "DGNUNDERLAY"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(11, 4.0),
		NewDoubleCodePair(21, 5.0),
		NewDoubleCodePair(11, 6.0),
		NewDoubleCodePair(21, 7.0),
	}, actual)
}

func TestReadWipeout(t *testing.T) {
	wo := parseEntity(t, "WIPEOUT",
		NewIntCodePair(91, 2),
		NewDoubleCodePair(14, 1.0),
		NewDoubleCodePair(24, 2.0),
		NewDoubleCodePair(14, 3.0),
		NewDoubleCodePair(24, 4.0),
	).(*Wipeout)
	assertEqInt(t, 2, len(wo.ClippingVertices()))
	assertEqPoint(t, Point{1.0, 2.0, 0.0}, wo.ClippingVertices()[0])
	assertEqPoint(t, Point{3.0, 4.0, 0.0}, wo.ClippingVertices()[1])
}

func TestWriteWipeout(t *testing.T) {
	wo := NewWipeout()
	wo.SetClippingVertices(append(wo.ClippingVertices(), Point{1.0, 2.0, 0.0}))
	wo.SetClippingVertices(append(wo.ClippingVertices(), Point{3.0, 4.0, 0.0}))
	actual := allCodePairs(wo, R2000)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(100, "AcDbWipeout"),
	}, actual)
	assertContainsCodePairs(t, []CodePair{
		NewIntCodePair(91, 2),
		NewDoubleCodePair(14, 1.0),
		NewDoubleCodePair(24, 2.0),
		NewDoubleCodePair(14, 3.0),
		NewDoubleCodePair(24, 4.0),
	}, actual)
}

func TestReadXLine(t *testing.T) {
	x := parseEntity(t, "XLINE",
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(11, 4.0),
		NewDoubleCodePair(21, 5.0),
		NewDoubleCodePair(31, 6.0),
	).(*XLine)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, x.FirstPoint)
	assertEqVector(t, Vector{4.0, 5.0, 6.0}, x.UnitDirectionVector)
}

func TestWriteXLine(t *testing.T) {
	x := NewXLine()
	x.FirstPoint = Point{1.0, 2.0, 3.0}
	x.UnitDirectionVector = Vector{4.0, 5.0, 6.0}
	actual := allCodePairs(x, R13)
	assertContainsCodePairs(t, []CodePair{
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
		NewDoubleCodePair(11, 4.0),
		NewDoubleCodePair(21, 5.0),
		NewDoubleCodePair(31, 6.0),
	}, actual)
}

func parseEntity(t *testing.T, entityType string, body ...CodePair) Entity {
	codePairs := []CodePair{NewStringCodePair(0, entityType)}
	codePairs = append(codePairs, body...)
	entities := parseEntities(t, codePairs...)
	assertEqInt(t, 1, len(entities))
	return entities[0]
}

func parseEntities(t *testing.T, body ...CodePair) []Entity {
	codePairs := []CodePair{
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "ENTITIES"),
	}
	codePairs = append(codePairs, body...)
	codePairs = append(codePairs,
		NewStringCodePair(0, "ENDSEC"),
		NewStringCodePair(0, "EOF"),
	)
	drawing := parseFromCodePairs(t, codePairs...)
	return drawing.Entities
}
