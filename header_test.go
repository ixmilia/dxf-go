package dxf

import (
	"fmt"
	"testing"
)

func TestReadNonDefaultHeaderVersion(t *testing.T) {
	header := parseHeader(t,
		NewStringCodePair(9, "$UNSUPPORTED_HEADER_VARIABLE"),
		NewStringCodePair(1, "UNSUPPORTED_VALUE"),
		NewStringCodePair(9, "$ACADVER"),
		NewStringCodePair(1, "AC1014"),
		NewStringCodePair(9, "$ACADMAINTVER"),
		NewShortCodePair(70, 6),
		NewStringCodePair(9, "$UNSUPPORTED_HEADER_VARIABLE"),
		NewStringCodePair(1, "UNSUPPORTED_VALUE"),
	)
	assertEqInt(t, int(R14), int(header.Version))
	assertEqInt(t, 6, int(header.MaintenanceVersion))
}

func TestWriteVersionSpecificVariables(t *testing.T) {
	header := *NewHeader()

	// value is present >= R14
	header.Version = R14
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(9, "$ACADMAINTVER"),
	}, fileCodePairsFromHeader(t, header))

	// value is missing < R14
	header.Version = R13
	assertNotContainsCodePairs(t, []CodePair{
		NewStringCodePair(9, "$ACADMAINTVER"),
	}, fileCodePairsFromHeader(t, header))
}

func TestReadHeaderFlag(t *testing.T) {
	header := parseHeader(t,
		NewStringCodePair(9, "$OSMODE"),
		NewShortCodePair(70, 1),
	)
	assert(t, header.EndPointSnap(), "expected OSMODE.EndPointSnap")
	assert(t, !header.MidPointSnap(), "expected !OSMODE.MidPointSnap")
	assert(t, !header.CenterSnap(), "expected !OSMODE.CenterSnap")
	assert(t, !header.NodeSnap(), "expected !OSMODE.NodeSnap")
	assert(t, !header.QuadrantSnap(), "expected !OSMODE.QuadrantSnap")
	assert(t, !header.IntersectionSnap(), "expected !OSMODE.IntersectionSnap")
	assert(t, !header.InsertionSnap(), "expected !OSMODE.InsertionSnap")
	assert(t, !header.PerpendicularSnap(), "expected !OSMODE.PerpendicularSnap")
	assert(t, !header.TangentSnap(), "expected !OSMODE.TangentSnap")
	assert(t, !header.NearestSnap(), "expected !OSMODE.NearestSnap")
	assert(t, !header.ApparentIntersectionSnap(), "expected !OSMODE.ApparentIntersectionSnap")
	assert(t, !header.ExtensionSnap(), "expected !OSMODE.ExtensionSnap")
	assert(t, !header.ParallelSnap(), "expected !OSMODE.ParallelSnap")
}

func TestWriteHeaderFlag(t *testing.T) {
	header := *NewHeader()
	header.SetEndPointSnap(true)
	header.SetMidPointSnap(false)
	header.SetCenterSnap(false)
	header.SetNodeSnap(false)
	header.SetQuadrantSnap(false)
	header.SetIntersectionSnap(false)
	header.SetInsertionSnap(false)
	header.SetPerpendicularSnap(false)
	header.SetTangentSnap(false)
	header.SetNearestSnap(false)
	header.SetApparentIntersectionSnap(false)
	header.SetExtensionSnap(false)
	header.SetParallelSnap(false)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(9, "$OSMODE"),
		NewShortCodePair(70, 1),
	}, fileCodePairsFromHeader(t, header))
}

func TestReadPoint(t *testing.T) {
	header := parseHeader(t,
		NewStringCodePair(9, "$INSBASE"),
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
	)
	expected := Point{1.0, 2.0, 3.0}
	assert(t, header.InsertionBase == expected, fmt.Sprintf("expected %s, got %s", expected.String(), header.InsertionBase.String()))
}

func TestWritePoint(t *testing.T) {
	header := *NewHeader()
	header.InsertionBase = Point{1.0, 2.0, 3.0}
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(9, "$INSBASE"),
		NewDoubleCodePair(10, 1.0),
		NewDoubleCodePair(20, 2.0),
		NewDoubleCodePair(30, 3.0),
	}, fileCodePairsFromHeader(t, header))
}

func TestReadEnumValue(t *testing.T) {
	header := parseHeader(t,
		NewStringCodePair(9, "$DRAGMODE"),
		NewShortCodePair(70, 2),
	)
	assertEqShort(t, int16(2), int16(header.DragMode))
}

func TestWriteEnumValue(t *testing.T) {
	header := *NewHeader()
	header.DragMode = DragModeAuto
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(9, "$DRAGMODE"),
		NewShortCodePair(70, 2),
	}, fileCodePairsFromHeader(t, header))
}

func TestReadHandleValue(t *testing.T) {
	header := parseHeader(t,
		NewStringCodePair(9, "$DRAGVS"),
		NewStringCodePair(349, "FF"),
	)
	assertEqUInt(t, uint32(255), uint32(header.SolidVisualStylePointer))
}

func TestWriteHandleValue(t *testing.T) {
	header := *NewHeader()
	header.Version = R2007 // min version R2007
	header.SolidVisualStylePointer = Handle(255)
	assertContainsCodePairs(t, []CodePair{
		NewStringCodePair(9, "$DRAGVS"),
		NewStringCodePair(349, "FF"),
	}, fileCodePairsFromHeader(t, header))
}

func parseHeader(t *testing.T, codePairs ...CodePair) Header {
	allPairs := []CodePair{
		NewStringCodePair(0, "SECTION"),
		NewStringCodePair(2, "HEADER"),
	}
	allPairs = append(allPairs, codePairs...)
	allPairs = append(allPairs,
		NewStringCodePair(0, "ENDSEC"),
		NewStringCodePair(0, "EOF"),
	)
	drawing := parseFromCodePairs(t, allPairs...)
	return drawing.Header
}

func fileCodePairsFromHeader(t *testing.T, h Header) []CodePair {
	drawing := *NewDrawing()
	drawing.Header = h
	codePairs, err := drawing.CodePairs()
	if err != nil {
		t.Error(err)
	}

	return codePairs
}
