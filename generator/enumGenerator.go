package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type xmlEnums struct {
	XMLName xml.Name  `xml:"Enums"`
	Enums   []xmlEnum `xml:"Enum"`
}

type xmlEnum struct {
	XMLName  xml.Name       `xml:"Enum"`
	Name     string         `xml:"Name,attr"`
	BaseType string         `xml:"BaseType,attr"`
	Values   []xmlEnumValue `xml:"Value"`
}

type xmlEnumValue struct {
	XMLName xml.Name `xml:"Value"`
	Name    string   `xml:"Name,attr"`
	Value   string   `xml:"Value,attr"`
}

func generateEnums() {
	specPath := "spec/EnumSpec.xml"
	file, err := os.Open(specPath)
	check(err)

	defer file.Close()

	enums, err := readEnums(file)
	check(err)

	var builder strings.Builder
	builder.WriteString("// Code generated at build time; DO NOT EDIT.\n")
	builder.WriteString("\n")
	builder.WriteString("package dxf\n")
	builder.WriteString("\n")
	builder.WriteString("import (\n")
	builder.WriteString("	\"fmt\"\n")
	builder.WriteString(")\n")
	builder.WriteString("\n")

	for _, enum := range enums {
		// declaration
		baseType := "int16"
		if len(enum.BaseType) > 0 {
			baseType = enum.BaseType
		}
		builder.WriteString(fmt.Sprintf("type %s %s\n", enum.Name, baseType))
		builder.WriteString("\n")
		builder.WriteString("const (\n")
		for _, value := range enum.Values {
			var tail string
			if len(value.Value) > 0 {
				tail = fmt.Sprintf(" %s = %s", enum.Name, value.Value)
			}
			builder.WriteString(fmt.Sprintf("	%s%s%s\n", enum.Name, value.Name, tail))
		}
		builder.WriteString(")\n")
		builder.WriteString("\n")

		// `String()`
		builder.WriteString(fmt.Sprintf("func (this %s) String() string {\n", enum.Name))
		builder.WriteString("	switch this {\n")
		seenValues := make(map[string]bool)
		for _, value := range enum.Values {
			if !seenValues[value.Value] {
				seenValues[value.Value] = true
				builder.WriteString(fmt.Sprintf("	case %s%s:\n", enum.Name, value.Name))
				builder.WriteString(fmt.Sprintf("		return \"%s%s\"\n", enum.Name, value.Name))
			}
		}
		builder.WriteString("	default:\n")
		builder.WriteString(fmt.Sprintf("		return fmt.Sprintf(\"%%v\", %s(this))\n", enum.BaseType))
		builder.WriteString("	}\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")
	}

	writeFile("enums.generated.go", builder)
}

func readEnums(reader io.Reader) ([]xmlEnum, error) {
	var enums xmlEnums
	if err := xml.NewDecoder(reader).Decode(&enums); err != nil {
		return nil, err
	}

	return enums.Enums, nil
}
