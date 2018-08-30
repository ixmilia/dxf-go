package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type XMLHeader struct {
	XMLName   xml.Name            `xml:"Header"`
	Variables []XMLHeaderVariable `xml:"Variable"`
}

type XMLHeaderVariable struct {
	XMLName        xml.Name `xml:"Variable"`
	Name           string   `xml:"Name,attr"`
	Code           int      `xml:"Code,attr"`
	Type           string   `xml:"Type,attr"`
	FieldName      string   `xml:"Field,attr"`
	DefaultValue   string   `xml:"DefaultValue,attr"`
	WriteConverter string   `xml:"WriteConverter,attr"`
	Comment        string   `xml:"Comment,attr"`
}

func generateHeader() {
	specPath := "spec/HeaderSpec.xml"
	file, err := os.Open(specPath)
	check(err)

	defer file.Close()

	variables, err := ReadHeader(file)
	check(err)

	var content string
	content += "import (\n"
	content += ")\n"
	content += "\n"
	content += "type Header struct {\n"
	for _, variable := range variables {
		content += fmt.Sprintf("	// The $%s header variable.  %s\n", variable.Name, variable.Comment)
		content += fmt.Sprintf("	%s %s\n", variable.FieldName, variable.Type)
	}
	content += "}\n"
	content += "\n"

	content += "func NewHeader() *Header {\n"
	content += "	return &Header{\n"
	for _, variable := range variables {
		content += fmt.Sprintf("		%s: %s,\n", variable.FieldName, variable.DefaultValue)
	}
	content += "	}\n"
	content += "}\n"
	content += "\n"

	content += "func (h Header) writeHeader(writer CodePairWriter) error {\n"
	content += "	pairs := make([]CodePair, 0)\n"
	content += "	pairs = append(pairs, NewStringCodePair(0, \"SECTION\"))\n"
	content += "	pairs = append(pairs, NewStringCodePair(2, \"HEADER\"))\n"
	for _, variable := range variables {
		content += "\n"
		content += fmt.Sprintf("	// $%s\n", variable.Name)
		content += fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(9, \"$%s\"))\n", variable.Name)
		value := fmt.Sprintf("h.%s", variable.FieldName)
		if len(variable.WriteConverter) > 0 {
			value = strings.Replace(variable.WriteConverter, "%v", value, -1)
		}
		codeTypeName := CodeTypeName(variable.Code)
		content += fmt.Sprintf("	pairs = append(pairs, New%sCodePair(%d, %s))\n", codeTypeName, variable.Code, value)
	}
	content += "\n"
	content += "	pairs = append(pairs, NewStringCodePair(0, \"ENDSEC\"))\n"
	content += "	for _, pair := range pairs {\n"
	content += "		err := writer.writeCodePair(pair)\n"
	content += "		if err != nil {\n"
	content += "			return err\n"
	content += "		}\n"
	content += "	}\n"
	content += "	return nil\n"
	content += "}\n"
	writeFile("header.generated.go", content)
}

func ReadHeader(reader io.Reader) ([]XMLHeaderVariable, error) {
	var header XMLHeader
	if err := xml.NewDecoder(reader).Decode(&header); err != nil {
		return nil, err
	}

	return header.Variables, nil
}
