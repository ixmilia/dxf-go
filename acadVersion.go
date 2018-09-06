package dxf

// AcadVersion represents the minimum version of AutoCAD that is expected to be able to read the file.
type AcadVersion int

const (
	// Version1_0 corresponds to the value "MC0.0"
	Version1_0 AcadVersion = iota

	// Version1_2 corresponds to the value "AC1.2"
	Version1_2

	// Version1_40 corresponds to the value "AC1.40"
	Version1_40

	// Version2_05 corresponds to the value "AC1.50"
	Version2_05

	// Version2_10 corresponds to the value "AC2.10"
	Version2_10

	// Version2_21 corresponds to the value "AC2.21"
	Version2_21

	// Version2_22 corresponds to the value "AC2.22"
	Version2_22

	// Version2_5 corresponds to the value "AC1002"
	Version2_5

	// Version2_6 corresponds to the value "AC1003"
	Version2_6

	// R9 corresponds to the value "AC1004"
	R9

	// R10 corresponds to the value "AC1006"
	R10

	// R11 corresponds to the value "AC1009"
	R11

	// R12 corresponds to the value "AC1009"
	R12

	// R13 corresponds to the value "AC1012"
	R13

	// R14 corresponds to the value "AC1014"
	R14

	// R2000 corresponds to the value "AC1015"
	R2000

	// R2004 corresponds to the value "AC1018"
	R2004

	// R2007 corresponds to the value "AC1021"
	R2007

	// R2010 corresponds to the value "AC1024"
	R2010

	// R2013 corresponds to the value "AC1027"
	R2013

	// R2018 corresponds to the value "AC1032"
	R2018
)

func parseAcadVersion(val string) AcadVersion {
	switch val {
	case "MC0.0":
		return Version1_0
	case "AC1.2":
		return Version1_2
	case "AC1.40":
		return Version1_40
	case "AC1.50":
		return Version2_05
	case "AC2.10":
		return Version2_10
	case "AC2.21":
		return Version2_21
	case "AC2.22":
		return Version2_22
	case "AC1002":
		return Version2_5
	case "AC1003":
		return Version2_6
	case "AC1004":
		return R9
	case "AC1006":
		return R10
	case "AC1009":
		// also R11
		return R12
	case "AC1012":
		return R13
	case "AC1014":
		return R14
	case "AC1015":
		return R2000
	case "AC1018":
		return R2004
	case "AC1021":
		return R2007
	case "AC1024":
		return R2010
	case "AC1027":
		return R2013
	case "AC1032":
		return R2018
	default:
		// TODO: add error handling?
		return R12
	}
}

func (v AcadVersion) String() string {
	switch v {
	case Version1_0:
		return "MC0.0"
	case Version1_2:
		return "AC1.2"
	case Version1_40:
		return "AC1.40"
	case Version2_05:
		return "AC1.50"
	case Version2_10:
		return "AC2.10"
	case Version2_21:
		return "AC2.21"
	case Version2_22:
		return "AC2.22"
	case Version2_5:
		return "AC1002"
	case Version2_6:
		return "AC1003"
	case R9:
		return "AC1004"
	case R10:
		return "AC1006"
	case R11:
		return "AC1009"
	case R12:
		return "AC1009"
	case R13:
		return "AC1012"
	case R14:
		return "AC1014"
	case R2000:
		return "AC1015"
	case R2004:
		return "AC1018"
	case R2007:
		return "AC1021"
	case R2010:
		return "AC1024"
	case R2013:
		return "AC1027"
	case R2018:
		return "AC1032"
	default:
		return "UNKNOWN"
	}
}
