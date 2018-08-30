package dxf

type AcadVersion int

const (
	Version_1_0 AcadVersion = iota
	Version_1_2
	Version_1_40
	Version_2_05
	Version_2_10
	Version_2_21
	Version_2_22
	Version_2_5
	Version_2_6
	R9
	R10
	R11
	R12
	R13
	R14
	R2000
	R2004
	R2007
	R2010
	R2013
	R2018
)

func (v AcadVersion) String() string {
	switch v {
	case Version_1_0:
		return "MC0.0"
	case Version_1_2:
		return "AC1.2"
	case Version_1_40:
		return "AC1.40"
	case Version_2_05:
		return "AC1.50"
	case Version_2_10:
		return "AC2.10"
	case Version_2_21:
		return "AC2.21"
	case Version_2_22:
		return "AC2.22"
	case Version_2_5:
		return "AC1002"
	case Version_2_6:
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
