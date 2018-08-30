package main

import (
	"os"
)

func main() {
	generateHeader()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeFile(path, content string) {
	f, err := os.Create(path)
	check(err)

	content = `// Code generated at build time; DO NOT EDIT.

package dxf

` + content
	_, err = f.WriteString(content)
	check(err)

	err = f.Close()
	check(err)
}
