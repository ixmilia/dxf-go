package main

import (
	"io/ioutil"
	"strings"
)

func main() {
	buf, err := ioutil.ReadFile("codePairHelper.go")
	check(err)
	content := string(buf)
	content = strings.Replace(content, "package dxf", "package main", 1)
	buf = []byte(content)
	err = ioutil.WriteFile("generator/codePairHelper.go", buf, 0644)
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
