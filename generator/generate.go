package main

import (
	"os"
	"strings"
)

func main() {
	generateEnums()
	generateHeader()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeFile(path string, content strings.Builder) {
	f, err := os.Create(path)
	check(err)

	_, err = f.WriteString(content.String())
	check(err)

	err = f.Close()
	check(err)
}
