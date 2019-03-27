package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type xmlTables struct {
	XMLName xml.Name   `xml:"Tables"`
	Tables  []xmlTable `xml:"Table"`
}

type xmlTable struct {
	XMLName    xml.Name       `xml:"Table"`
	Collection string         `xml:"Collection,attr"`
	TypeString string         `xml:"TypeString,attr"`
	Items      []xmlTableItem `xml:"TableItem"`
}

type xmlTableItem struct {
	XMLName   xml.Name   `xml:"TableItem"`
	Name      string     `xml:"Name,attr"`
	ClassName string     `xml:"ClassName,attr"`
	Fields    []xmlField `xml:"Field"`
}

func generateTables() {
	specPath := "spec/TableSpec.xml"
	file, err := os.Open(specPath)
	check(err)

	defer file.Close()

	tables, err := readTables(file)
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

	// item
	for _, table := range tables {
		tableItem := table.Items[0]
		// declaration
		seenFields := make(map[string]bool)
		builder.WriteString(fmt.Sprintf("type %s struct {\n", tableItem.Name))
		for _, field := range tableItem.Fields {
			if !seenFields[field.Name] {
				seenFields[field.Name] = true
				builder.WriteString(fmt.Sprintf("	%s %s\n", field.Name, field.Type))
			}
		}
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// constructor
		seenFields = make(map[string]bool)
		builder.WriteString(fmt.Sprintf("func New%s() *%s {\n", tableItem.Name, tableItem.Name))
		builder.WriteString(fmt.Sprintf("	return &%s{\n", tableItem.Name))
		for _, field := range tableItem.Fields {
			if !seenFields[field.Name] {
				seenFields[field.Name] = true
				builder.WriteString(fmt.Sprintf("		%s: %s,\n", field.Name, field.DefaultValue))
			}
		}
		builder.WriteString("	}\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// reader
		builder.WriteString(fmt.Sprintf("func read%s(drawing *Drawing, np CodePair, reader codePairReader) (nextPair CodePair, error error) {\n", table.Collection))
		builder.WriteString("	nextPair = np\n")
		builder.WriteString("	for error == nil && !nextPair.isEndTable() {\n")
		builder.WriteString(fmt.Sprintf("		if nextPair.Code != 0 || nextPair.Value.(StringCodePairValue).Value != \"%s\" {\n", table.TypeString))
		builder.WriteString(fmt.Sprintf("			error = errors.New(\"expected 0/%s\")\n", table.TypeString))
		builder.WriteString("			return\n")
		builder.WriteString("		}\n")
		builder.WriteString(fmt.Sprintf("		item := *New%s()\n", tableItem.Name))
		builder.WriteString("		nextPair, error = reader.readCodePair()\n")
		builder.WriteString("		for error == nil && nextPair.Code != 0 {\n")
		builder.WriteString("			item.tryApplyCodePair(nextPair)\n")
		builder.WriteString("			nextPair, error = reader.readCodePair()\n")
		builder.WriteString("		}\n")
		builder.WriteString(fmt.Sprintf("		drawing.%s = append(drawing.%s, item)\n", table.Collection, table.Collection))
		builder.WriteString("	}\n")
		builder.WriteString("	return\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// tryApplyCodePair
		generateReader := true
		if generateReader {
			builder.WriteString(fmt.Sprintf("func (this *%s) tryApplyCodePair(codePair CodePair) {\n", tableItem.Name))
			builder.WriteString("	switch codePair.Code {\n")
			for _, field := range tableItem.Fields {
				readField(&builder, field, false)
			}
			builder.WriteString("	default:\n")
			builder.WriteString("	}\n")
			builder.WriteString("}\n")
			builder.WriteString("\n")
		}

		// writer
		builder.WriteString(fmt.Sprintf("func tablePairs%s(items []%s, version AcadVersion) (pairs []CodePair) {\n", table.Collection, tableItem.Name))
		builder.WriteString("	pairs = append(pairs, NewStringCodePair(0, \"TABLE\"))\n")
		builder.WriteString("	pairs = append(pairs, NewStringCodePair(100, \"AcDbSymbolTable\"))\n")
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(2, \"%s\"))\n", table.TypeString))
		builder.WriteString(fmt.Sprintf("	for _, item := range items {\n"))
		builder.WriteString(fmt.Sprintf("		pairs = append(pairs, NewStringCodePair(0, \"%s\"))\n", table.TypeString))
		builder.WriteString("		pairs = append(pairs, item.codePairs(version)...)\n")
		builder.WriteString("	}\n")
		builder.WriteString("	pairs = append(pairs, NewStringCodePair(0, \"ENDTAB\"))\n")
		builder.WriteString("	return\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// codePairs
		builder.WriteString(fmt.Sprintf("func (this *%s) codePairs(version AcadVersion) (pairs []CodePair) {\n", tableItem.Name))
		builder.WriteString("	pairs = append(pairs, NewStringCodePair(100, \"AcDbSymbolTableRecord\"))\n")
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(100, \"%s\"))\n", tableItem.ClassName))
		for _, field := range tableItem.Fields {
			writeField(&builder, field, false, "")
		}
		builder.WriteString("	return\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")
	}

	// general reader
	builder.WriteString("func readSpecificTable(drawing *Drawing, np CodePair, reader codePairReader, tableType string) (nextPair CodePair, error error) {\n")
	builder.WriteString("	nextPair = np\n")
	builder.WriteString("	if error != nil {\n")
	builder.WriteString("		return\n")
	builder.WriteString("	}\n")
	builder.WriteString("	switch tableType {\n")
	for _, table := range tables {
		builder.WriteString(fmt.Sprintf("	case \"%s\":\n", table.TypeString))
		builder.WriteString(fmt.Sprintf("		nextPair, error = read%s(drawing, nextPair, reader)\n", table.Collection))
	}
	builder.WriteString("	}\n")
	builder.WriteString("	return\n")
	builder.WriteString("}\n")
	builder.WriteString("\n")

	// general writer
	builder.WriteString("func getTablePairs(drawing *Drawing, version AcadVersion) (pairs []CodePair) {\n")
	for _, table := range tables {
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, tablePairs%s(drawing.%s, version)...)\n", table.Collection, table.Collection))
	}
	builder.WriteString("	return\n")
	builder.WriteString("}\n")

	writeFile("tables.generated.go", builder)
}

func readTables(reader io.Reader) ([]xmlTable, error) {
	var tables xmlTables
	if err := xml.NewDecoder(reader).Decode(&tables); err != nil {
		return nil, err
	}

	return tables.Tables, nil
}
