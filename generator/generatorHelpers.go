package main

import "fmt"

func generateComment(mainComment, minVersion, maxVersion string) string {
	comment := mainComment
	if len(minVersion) > 0 {
		comment += fmt.Sprintf("  Minimum AutoCAD version %s.", minVersion)
	}
	if len(maxVersion) > 0 {
		comment += fmt.Sprintf("  Maximum AutoCAD version %s.", maxVersion)
	}
	return comment
}
