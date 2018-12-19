package dxf

import (
	"strings"
	"testing"
)

func TestParseSimpleLine(t *testing.T) {
	line := parseEntity(t, "LINE", join(
		" 10", "1.0",
		" 20", "2.0",
		" 30", "3.0",
		" 11", "4.0",
		" 21", "5.0",
		" 31", "6.0",
	)).(*Line)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, line.P1)
	assertEqPoint(t, Point{4.0, 5.0, 6.0}, line.P2)
}

func TestParseAlternateTypeString(t *testing.T) {
	line := parseEntity(t, "3DLINE", join(
		" 10", "1.0",
		" 20", "2.0",
		" 30", "3.0",
		" 11", "4.0",
		" 21", "5.0",
		" 31", "6.0",
	)).(*Line)
	assertEqPoint(t, Point{1.0, 2.0, 3.0}, line.P1)
	assertEqPoint(t, Point{4.0, 5.0, 6.0}, line.P2)
}

func TestParseUnsupportedEntity(t *testing.T) {
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "ENTITIES",
		"  0", "LINE", // supported entity
		"  0", "NOT_AN_ENTITY", // unsupported entity
		"  0", "LINE", // supported entity
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	assertEqInt(t, 2, len(drawing.Entities))
	assertEqString(t, "LINE", drawing.Entities[0].typeString())
	assertEqString(t, "LINE", drawing.Entities[1].typeString())
}

func TestWriteSimpleLine(t *testing.T) {
	line := NewLine()
	line.P1 = Point{1.0, 2.0, 3.0}
	line.P2 = Point{4.0, 5.0, 6.0}
	actual := entityString(line, R12)
	assertContains(t, join(
		"  0", "LINE",
	), actual)
	assertContains(t, join(
		" 10", "1.0",
		" 20", "2.0",
		" 30", "3.0",
		" 11", "4.0",
		" 21", "5.0",
		" 31", "6.0",
	), actual)
}

func TestConditionalEntityFieldWriting(t *testing.T) {
	line := NewLine()
	line.SetIsInPaperSpace(false)
	actual := entityString(line, R14)
	assertNotContains(t, join(
		"  0", "LINE",
		"100", "AcDbEntity",
		" 67", // [NO-VALUE] this is only written when Version >= R12 and it's not the default (false)
	), actual)
}

func TestReadEntityFieldFlag(t *testing.T) {
	face := parseEntity(t, "3DFACE", join(
		" 70", "     5",
	)).(*Face)
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
	actual := entityString(face, R12)
	assertContains(t, join(
		// these parts just ensure we're checking the correct entity
		"100", "AcDbFace",
		" 10", "0.0",
		" 20", "0.0",
		" 30", "0.0",
		" 11", "0.0",
		" 21", "0.0",
		" 31", "0.0",
		" 12", "0.0",
		" 22", "0.0",
		" 32", "0.0",
		" 13", "0.0",
		" 23", "0.0",
		" 33", "0.0",
		// this is the real check
		" 70", "     5",
	), actual)
}

func TestWriteVersionSpecificEntities(t *testing.T) {
	solid := NewSolid()
	drawing := *NewDrawing()
	drawing.Entities = append(drawing.Entities, solid)

	// ensure it's present when appropriate
	drawing.Header.Version = R13
	assertContains(t, join(
		"  0", "3DSOLID",
	), drawing.String())

	// and not otherwise
	drawing.Header.Version = R12
	assertNotContains(t, join(
		"  0", "3DSOLID",
	), drawing.String())
}

func TestReadMultipleBaseEntityData(t *testing.T) {
	line := parseEntity(t, "LINE", join(
		"310", "line 1",
		"310", "line 2",
	)).(*Line)
	assertEqInt(t, 2, len(line.PreviewImageData()))
	assertEqString(t, "line 1", line.PreviewImageData()[0])
	assertEqString(t, "line 2", line.PreviewImageData()[1])
}

func TestWriteMultipleBaseEntityData(t *testing.T) {
	line := NewLine()
	line.SetPreviewImageData(append(line.PreviewImageData(), "line 1"))
	line.SetPreviewImageData(append(line.PreviewImageData(), "line 2"))
	actual := entityString(line, R2000)
	assertContains(t, join(
		"310", "line 1",
		"310", "line 2",
		"100", "AcDbLine",
	), actual)
}

func TestReadMultipleSpecificEntityData(t *testing.T) {
	solid := parseEntity(t, "3DSOLID", join(
		"  1", "line 1",
		"  1", "line 2",
	)).(*Solid)
	assertEqInt(t, 2, len(solid.CustomData))
	assertEqString(t, "line 1", solid.CustomData[0])
	assertEqString(t, "line 2", solid.CustomData[1])
}

func TestWriteMultipleSpecificEntityData(t *testing.T) {
	solid := NewSolid()
	solid.AddCustomData("line 1")
	solid.AddCustomData("line 2")
	actual := entityString(solid, R13)
	assertContains(t, join(
		"100", "AcDbModelerGeometry",
		" 70", "     1",
		"  1", "line 1",
		"  1", "line 2",
	), actual)
}

func TestWriteConditionsOnWriteOrderDirectives(t *testing.T) {
	solid := NewSolid()
	solid.AddCustomData("custom data")

	actual := entityString(solid, R2007)
	assertContains(t, join(
		"100", "AcDb3dSolid",
	), actual)

	actual = entityString(solid, R13)
	assertNotContains(t, join(
		"100", "AcDb3dSolid",
	), actual)
}

func TestReadEntityWithCustomReader(t *testing.T) {
	proxy := parseEntity(t, "ACAD_PROXY_ENTITY", join(
		" 92", "4",
		"310", "1234",
		"310", "ABCD",
		" 93", "4",
		"310", "5678",
		"310", "DCBA",
	)).(*ProxyEntity)
	assertEqByteArray(t, []byte{0x12, 0x34, 0xAB, 0xCD}, proxy.GraphicsData)
	assertEqByteArray(t, []byte{0x56, 0x78, 0xDC, 0xBA}, proxy.EntityData)
}

func TestWriteEntityWithBeforeWrite(t *testing.T) {
	proxy := NewProxyEntity()
	proxy.GraphicsData = []byte{0x12, 0x34, 0xAB, 0xCD}
	proxy.EntityData = []byte{0x56, 0x78, 0xDC, 0xBA}
	actual := entityString(proxy, R14)
	assertContains(t, join(
		" 92", "        4",
		"310", "1234ABCD",
		" 93", "        4",
		"310", "5678DCBA",
	), actual)
}

func TestReadCollectedEntities(t *testing.T) {
	attdef := parseEntity(t, "ATTDEF", join(
		"  0", "MTEXT",
		"  1", "mtext-value",
	)).(*AttributeDefinition)
	assertEqString(t, "mtext-value", attdef.MText.Text)
}

func TestWriteEntityWithTrailingEntities(t *testing.T) {
	attdef := NewAttributeDefinition()
	actual := entityString(attdef, R14)
	assertContains(t, "\r\n  0\r\nATTDEF\r\n", actual)
	assertContains(t, "\r\n  0\r\nMTEXT\r\n", actual)
}

func TestReadDimension(t *testing.T) {
	dim := parseEntity(t, "DIMENSION", join(
		"  1", "text",
		" 10", "1.0",
		" 20", "2.0",
		" 70", "1", // aligned
		"100", "AcDbAlignedDimension",
		" 13", "3.0",
		" 23", "4.0",
		" 14", "5.0",
		" 24", "6.0",
	)).(*AlignedDimension)
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
	actual := entityString(dim, R14)
	assertContains(t, join(
		" 10", "1.0",
		" 20", "2.0",
		" 30", "0.0",
		" 11", "0.0",
		" 21", "0.0",
		" 31", "0.0",
		" 70", "     1",
	), actual)
	assertContains(t, join(
		"100", "AcDbAlignedDimension",
		" 13", "3.0",
		" 23", "4.0",
		" 33", "0.0",
		" 14", "5.0",
		" 24", "6.0",
		" 34", "0.0",
	), actual)
}

func TestReadImage(t *testing.T) {
	img := parseEntity(t, "IMAGE", join(
		" 91", "2",
		" 14", "1.0",
		" 24", "2.0",
		" 14", "3.0",
		" 24", "4.0",
	)).(*Image)
	assertEqInt(t, 2, len(img.ClippingVertices))
	assertEqPoint(t, Point{1.0, 2.0, 0.0}, img.ClippingVertices[0])
	assertEqPoint(t, Point{3.0, 4.0, 0.0}, img.ClippingVertices[1])
}

func TestWriteImage(t *testing.T) {
	img := NewImage()
	img.ClippingVertices = append(img.ClippingVertices, Point{1.0, 2.0, 0.0})
	img.ClippingVertices = append(img.ClippingVertices, Point{3.0, 4.0, 0.0})
	actual := entityString(img, R14)
	assertContains(t, join(
		" 91", "        2",
		" 14", "1.0",
		" 24", "2.0",
		" 14", "3.0",
		" 24", "4.0",
	), actual)
}

func TestReadInsertAtEnd(t *testing.T) {
	ins := parseEntity(t, "INSERT", join(
		" 66", "1", // has attributes
		"  0", "ATTRIB",
		"  1", "attrib 1",
		"  0", "ATTRIB",
		"  1", "attrib 2",
		"  0", "SEQEND",
	)).(*Insert)
	assertEqInt(t, 2, len(ins.Attributes))
	assertEqString(t, "attrib 1", ins.Attributes[0].Value)
	assertEqString(t, "attrib 2", ins.Attributes[1].Value)
}

func TestReadInsertAtEndNoSeqend(t *testing.T) {
	ins := parseEntity(t, "INSERT", join(
		" 66", "1", // has attributes
		"  0", "ATTRIB",
		"  1", "attrib 1",
		"  0", "ATTRIB",
		"  1", "attrib 2",
	)).(*Insert)
	assertEqInt(t, 2, len(ins.Attributes))
	assertEqString(t, "attrib 1", ins.Attributes[0].Value)
	assertEqString(t, "attrib 2", ins.Attributes[1].Value)
}

func TestReadInsertWithTrailingEntity(t *testing.T) {
	entities := parseEntities(t, join(
		"  0", "INSERT",
		" 66", "1", // has attributes
		"  0", "ATTRIB",
		"  1", "attrib 1",
		"  0", "ATTRIB",
		"  1", "attrib 2",
		"  0", "SEQEND",
		"  0", "LINE", // trailing entity
		" 10", "111.1",
	))
	assertEqInt(t, 2, len(entities))
	ins := entities[0].(*Insert)
	assertEqInt(t, 2, len(ins.Attributes))
	assertEqString(t, "attrib 1", ins.Attributes[0].Value)
	assertEqString(t, "attrib 2", ins.Attributes[1].Value)
	line := entities[1].(*Line)
	assertEqPoint(t, Point{111.1, 0.0, 0.0}, line.P1)
}

func TestReadInsertWithTrailingEntityNoSeqend(t *testing.T) {
	entities := parseEntities(t, join(
		"  0", "INSERT",
		" 66", "1", // has attributes
		"  0", "ATTRIB",
		"  1", "attrib 1",
		"  0", "ATTRIB",
		"  1", "attrib 2",
		"  0", "LINE", // trailing entity
		" 10", "111.1",
	))
	assertEqInt(t, 2, len(entities))
	ins := entities[0].(*Insert)
	assertEqInt(t, 2, len(ins.Attributes))
	assertEqString(t, "attrib 1", ins.Attributes[0].Value)
	assertEqString(t, "attrib 2", ins.Attributes[1].Value)
	line := entities[1].(*Line)
	assertEqPoint(t, Point{111.1, 0.0, 0.0}, line.P1)
}

func TestWriteInsert(t *testing.T) {
	ins := NewInsert()
	att1 := *NewAttribute()
	att1.Value = "attrib 1"
	ins.Attributes = append(ins.Attributes, att1)
	att2 := *NewAttribute()
	att2.Value = "attrib 2"
	ins.Attributes = append(ins.Attributes, att2)
	actual := entityString(ins, R14)
	assertContains(t, join(
		"  1", "attrib 1",
	), actual)
	assertContains(t, join(
		"  1", "attrib 2",
	), actual)
	assertContains(t, join(
		"  0", "SEQEND",
	), actual)
}

func parseEntity(t *testing.T, entityType string, body string) Entity {
	entities := parseEntities(t, join(
		"  0", entityType,
	)+"\r\n"+strings.TrimSpace(body))
	assertEqInt(t, 1, len(entities))
	return entities[0]
}

func parseEntities(t *testing.T, body string) []Entity {
	drawing := parse(t, join(
		"  0", "SECTION",
		"  2", "ENTITIES",
	)+"\r\n"+strings.TrimSpace(body)+"\r\n"+join(
		"  0", "ENDSEC",
		"  0", "EOF",
	))
	return drawing.Entities
}

func entityString(entity Entity, version AcadVersion) string {
	drawing := *NewDrawing()
	drawing.Header.Version = version
	drawing.Entities = append(drawing.Entities, entity)
	return drawing.String()
}
