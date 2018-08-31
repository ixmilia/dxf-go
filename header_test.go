package dxf

import (
	"strings"
	"testing"
)

func TestReadNonDefaultHeaderVersion(t *testing.T) {
	header := parseHeader(t, `
  9
$UNSUPPORTED_HEADER_VARIABLE
  1
UNSUPPORTED_VALUE
  9
$ACADVER
  1
AC1014
  9
$ACADMAINTVER
 70
6
  9
$ANOTHER_UNSUPPORTED_HEADER_VARIABLE
  1
ANOTHER_UNSUPPORTED_VALUE
`)
	assertEqInt(t, int(R14), int(header.Version))
	assertEqInt(t, 6, int(header.MaintenanceVersion))
}

func TestWriteVersionSpecificVariables(t *testing.T) {
	header := *NewHeader()

	// value is present >= R14
	header.Version = R14
	assertContains(t, "$ACADMAINTVER", fileStringFromHeader(header))

	// value is missing < R14
	header.Version = R13
	assertNotContains(t, "$ACADMAINTVER", fileStringFromHeader(header))
}

func parseHeader(t *testing.T, content string) Header {
	drawing := parse(t, `
  0
SECTION
  2
HEADER
`+strings.TrimSpace(content)+`
  0
ENDSEC
  0
EOF
`)
	return drawing.Header
}

func TestReadHeaderFlag(t *testing.T) {
	header := parseHeader(t, `
  9
$OSMODE
 70
1
`)
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
	assertContains(t, "  9\r\n$OSMODE\r\n 70\r\n     1\r\n", fileStringFromHeader(header))
}

func fileStringFromHeader(h Header) string {
	drawing := *NewDrawing()
	drawing.Header = h
	return drawing.String()
}
