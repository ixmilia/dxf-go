package dxf

func codeTypeName(code int) string {
	// official code types
	if between(code, 0, 9) {
		return "String"
	} else if between(code, 10, 39) {
		return "Double"
	} else if between(code, 40, 59) {
		return "Double"
	} else if between(code, 60, 79) {
		return "Short"
	} else if between(code, 90, 99) {
		return "Int"
	} else if between(code, 100, 102) {
		return "String"
	} else if code == 105 {
		return "String"
	} else if between(code, 110, 119) {
		return "Double"
	} else if between(code, 120, 129) {
		return "Double"
	} else if between(code, 130, 139) {
		return "Double"
	} else if between(code, 140, 149) {
		return "Double"
	} else if between(code, 160, 169) {
		return "Long"
	} else if between(code, 170, 179) {
		return "Short"
	} else if between(code, 210, 239) {
		return "Double"
	} else if between(code, 270, 279) {
		return "Short"
	} else if between(code, 280, 289) {
		return "Short"
	} else if between(code, 290, 299) {
		return "Bool"
	} else if between(code, 300, 309) {
		return "String"
	} else if between(code, 310, 319) {
		return "String"
	} else if between(code, 320, 329) {
		return "String"
	} else if between(code, 330, 369) {
		return "String"
	} else if between(code, 370, 379) {
		return "Short"
	} else if between(code, 380, 389) {
		return "Short"
	} else if between(code, 390, 399) {
		return "String"
	} else if between(code, 400, 409) {
		return "Short"
	} else if between(code, 410, 419) {
		return "String"
	} else if between(code, 420, 429) {
		return "Int"
	} else if between(code, 430, 439) {
		return "String"
	} else if between(code, 440, 449) {
		return "Int"
	} else if between(code, 450, 459) {
		return "Int"
	} else if between(code, 460, 469) {
		return "Double"
	} else if between(code, 470, 479) {
		return "String"
	} else if between(code, 480, 481) {
		return "String"
	} else if code == 999 {
		return "String"
	} else if between(code, 1000, 1009) {
		return "String"
	} else if between(code, 1010, 1059) {
		return "Double"
	} else if between(code, 1060, 1070) {
		return "Short"
	} else if code == 1071 {
		return "Int"
	} else if code == 250 { // UNOFFICIAL: used in POLYLINEs by CLO
		return "Short"
	} else {
		return "Unknown"
	}
}

func between(val, lower, upper int) bool {
	return val >= lower && val <= upper
}
