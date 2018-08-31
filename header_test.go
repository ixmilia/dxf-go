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

func fileStringFromHeader(h Header) string {
	drawing := *NewDrawing()
	drawing.Header = h
	return drawing.String()
}
