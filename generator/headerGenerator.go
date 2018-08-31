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
	XMLName        xml.Name                `xml:"Variable"`
	Name           string                  `xml:"Name,attr"`
	Code           int                     `xml:"Code,attr"`
	Type           string                  `xml:"Type,attr"`
	FieldName      string                  `xml:"Field,attr"`
	DefaultValue   string                  `xml:"DefaultValue,attr"`
	ReadConverter  string                  `xml:"ReadConverter,attr"`
	WriteConverter string                  `xml:"WriteConverter,attr"`
	MinVersion     string                  `xml:"MinVersion,attr"`
	MaxVersion     string                  `xml:"MaxVersion,attr"`
	Comment        string                  `xml:"Comment,attr"`
	Flags          []XMLHeaderVariableFlag `xml:"Flag"`
}

type XMLHeaderVariableFlag struct {
	XMLName xml.Name `xml:"Flag"`
	Name    string   `xml:"Name,attr"`
	Mask    int      `xml:"Mask,attr"`
	Comment string   `xml:"Comment,attr"`
}

func generateHeader() {
	specPath := "spec/HeaderSpec.xml"
	file, err := os.Open(specPath)
	check(err)

	defer file.Close()

	variables, err := ReadHeader(file)
	check(err)

	var builder strings.Builder
	builder.WriteString("// Code generated at build time; DO NOT EDIT.\n")
	builder.WriteString("\n")
	builder.WriteString("package dxf\n")
	builder.WriteString("\n")
	builder.WriteString("import (\n")
	builder.WriteString("	\"errors\"\n")
	builder.WriteString(")\n")
	builder.WriteString("\n")

	// definition
	builder.WriteString("// Header contains the values common to an AutoCAD DXF drawing.\n")
	builder.WriteString("type Header struct {\n")
	for _, variable := range variables {
		comment := generateComment(fmt.Sprintf("The $%s header variable.  %s", variable.Name, variable.Comment), variable.MinVersion, variable.MaxVersion)
		builder.WriteString(fmt.Sprintf("	// %s\n", comment))
		builder.WriteString(fmt.Sprintf("	%s %s\n", variable.FieldName, variable.Type))
	}
	builder.WriteString("}\n")
	builder.WriteString("\n")

	// constructor
	builder.WriteString("func NewHeader() *Header {\n")
	builder.WriteString("	return &Header{\n")
	for _, variable := range variables {
		builder.WriteString(fmt.Sprintf("		%s: %s,\n", variable.FieldName, variable.DefaultValue))
	}
	builder.WriteString("	}\n")
	builder.WriteString("}\n")
	builder.WriteString("\n")

	// flags
	for _, variable := range variables {
		for _, flag := range variable.Flags {
			comment := generateComment(fmt.Sprintf("%s status flag.", flag.Name), variable.MinVersion, variable.MaxVersion)
			builder.WriteString(fmt.Sprintf("// %s\n", comment))
			builder.WriteString(fmt.Sprintf("func (h *Header) %s() (val bool) {\n", flag.Name))
			builder.WriteString(fmt.Sprintf("	return h.%s & %d != 0\n", variable.FieldName, flag.Mask))
			builder.WriteString("}\n")
			builder.WriteString("\n")

			builder.WriteString(fmt.Sprintf("// %s\n", comment))
			builder.WriteString(fmt.Sprintf("func (h *Header) Set%s(val bool) {\n", flag.Name))
			builder.WriteString("	if val {\n")
			builder.WriteString(fmt.Sprintf("		h.%s = h.%s | %d\n", variable.FieldName, variable.FieldName, flag.Mask))
			builder.WriteString("	} else {\n")
			builder.WriteString(fmt.Sprintf("		h.%s = h.%s & ^%d\n", variable.FieldName, variable.FieldName, flag.Mask))
			builder.WriteString("	}\n")
			builder.WriteString("}\n")
			builder.WriteString("\n")
		}
	}

	// writeHeader()
	builder.WriteString("func (h Header) writeHeader(writer CodePairWriter) error {\n")
	builder.WriteString("	pairs := make([]CodePair, 0)\n")
	builder.WriteString("	pairs = append(pairs, NewStringCodePair(0, \"SECTION\"))\n")
	builder.WriteString("	pairs = append(pairs, NewStringCodePair(2, \"HEADER\"))\n")
	for _, variable := range variables {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("	// $%s\n", variable.Name))
		var predicates []string
		if len(variable.MinVersion) > 0 {
			predicates = append(predicates, fmt.Sprintf("h.Version >= %s", variable.MinVersion))
		}
		if len(variable.MaxVersion) > 0 {
			predicates = append(predicates, fmt.Sprintf("h.Version <= %s", variable.MaxVersion))
		}
		indention := ""
		if len(predicates) > 0 {
			indention = "	"
			builder.WriteString(fmt.Sprintf("	if %s {\n", strings.Join(predicates, " && ")))
		}
		builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, NewStringCodePair(9, \"$%s\"))\n", indention, variable.Name))
		value := fmt.Sprintf("h.%s", variable.FieldName)
		if len(variable.WriteConverter) > 0 {
			value = strings.Replace(variable.WriteConverter, "%v", value, -1)
		}
		codeTypeName := CodeTypeName(variable.Code)
		builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, New%sCodePair(%d, %s))\n", indention, codeTypeName, variable.Code, value))
		if len(predicates) > 0 {
			builder.WriteString(fmt.Sprintf("	}\n"))
		}
	}
	builder.WriteString("\n")
	builder.WriteString("	pairs = append(pairs, NewStringCodePair(0, \"ENDSEC\"))\n")
	builder.WriteString("	for _, pair := range pairs {\n")
	builder.WriteString("		err := writer.writeCodePair(pair)\n")
	builder.WriteString("		if err != nil {\n")
	builder.WriteString("			return err\n")
	builder.WriteString("		}\n")
	builder.WriteString("	}\n")
	builder.WriteString("	return nil\n")
	builder.WriteString("}\n")
	builder.WriteString("\n")

	// readHeader()
	builder.WriteString("func readHeader(nextPair CodePair, reader CodePairReader) (Header, CodePair, error) {\n")
	builder.WriteString("	header := *NewHeader()\n")
	builder.WriteString("	var err error\n")
	builder.WriteString("	var variableName string\n")
	builder.WriteString("	for nextPair.Code != 0 {\n")

	// look for 9/$VARNAME
	builder.WriteString("		if nextPair.Code == 9 {\n")
	builder.WriteString("			variableName = nextPair.Value.(StringCodePairValue).Value\n")
	builder.WriteString("		} else {\n")
	builder.WriteString("			switch variableName {\n")
	for _, variable := range variables {
		builder.WriteString(fmt.Sprintf("			case \"$%s\":\n", variable.Name))
		// validate code
		builder.WriteString(fmt.Sprintf("				if nextPair.Code != %d {\n", variable.Code))
		builder.WriteString(fmt.Sprintf("					return header, nextPair, errors.New(\"expected code %d\")\n", variable.Code))
		builder.WriteString("				}\n")

		// read the value
		readValue := fmt.Sprintf("nextPair.Value.(%sCodePairValue).Value", CodeTypeName(variable.Code))
		if len(variable.ReadConverter) > 0 {
			readValue = strings.Replace(variable.ReadConverter, "%v", readValue, -1)
		}
		builder.WriteString(fmt.Sprintf("				header.%s = %s\n", variable.FieldName, readValue))
	}
	builder.WriteString("			default:\n")
	builder.WriteString("				// ignore unsupported header variable\n")
	builder.WriteString("			}\n")
	builder.WriteString("		}\n")
	builder.WriteString("\n")
	builder.WriteString("		nextPair, err = reader.readCodePair()\n")
	builder.WriteString("		if err != nil {\n")
	builder.WriteString("			return header, nextPair, err\n")
	builder.WriteString("		}\n")
	builder.WriteString("	}\n")
	builder.WriteString("\n")
	builder.WriteString("	return header, nextPair, nil\n")
	builder.WriteString("}\n")

	writeFile("header.generated.go", builder)
}

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

func ReadHeader(reader io.Reader) ([]XMLHeaderVariable, error) {
	var header XMLHeader
	if err := xml.NewDecoder(reader).Decode(&header); err != nil {
		return nil, err
	}

	return header.Variables, nil
}
