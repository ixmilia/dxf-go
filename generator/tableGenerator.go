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
	MinVersion string         `xml:"MinVersion,attr"`
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
	builder.WriteString("	\"fmt\"\n")
	builder.WriteString(")\n")
	builder.WriteString("\n")

	// item
	for _, table := range tables {
		tableItem := table.Items[0]
		// declaration
		seenFields := make(map[string]bool)
		builder.WriteString(fmt.Sprintf("type %s struct {\n", tableItem.Name))
		builder.WriteString("	handle Handle\n")
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

		// handles
		builder.WriteString(fmt.Sprintf("func (this *%s) Handle() Handle {\n", tableItem.Name))
		builder.WriteString("	return this.handle\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("func (this *%s) SetHandle(val Handle) {\n", tableItem.Name))
		builder.WriteString("	this.handle = val\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// reader
		builder.WriteString(fmt.Sprintf("func read%s(drawing *Drawing, np CodePair, reader codePairReader) (nextPair CodePair, error error) {\n", table.Collection))
		builder.WriteString("	nextPair = np\n")
		builder.WriteString("	for error == nil && !nextPair.isEndTable() {\n")
		builder.WriteString(fmt.Sprintf("		if nextPair.Code != 0 || nextPair.Value.(StringCodePairValue).Value != \"%s\" {\n", table.TypeString))
		builder.WriteString(fmt.Sprintf("			error = fmt.Errorf(\"expected [0/%s] but found [%%s]\", nextPair.String())\n", table.TypeString))
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
		builder.WriteString(fmt.Sprintf("func tablePairs%s(tableHandle Handle, items []%s, version AcadVersion) (pairs []CodePair) {\n", table.Collection, tableItem.Name))
		builder.WriteString("	pairs = append(pairs, NewStringCodePair(0, \"TABLE\"))\n")
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(2, \"%s\"))\n", table.TypeString))
		builder.WriteString("	pairs = append(pairs, NewStringCodePair(5, stringFromHandle(tableHandle)))\n")
		builder.WriteString("	if version >= R13 {\n")
		builder.WriteString("		pairs = append(pairs, NewStringCodePair(100, \"AcDbSymbolTable\"))\n")
		builder.WriteString("	}\n")
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewShortCodePair(70, int16(len(items))))\n"))
		builder.WriteString(fmt.Sprintf("	for _, item := range items {\n"))
		builder.WriteString(fmt.Sprintf("		pairs = append(pairs, NewStringCodePair(0, \"%s\"))\n", table.TypeString))
		handleCode := 5
		if table.TypeString == "DIMSTYLE" {
			handleCode = 105
		}
		builder.WriteString(fmt.Sprintf("		pairs = append(pairs, NewStringCodePair(%d, stringFromHandle(item.Handle())))\n", handleCode))
		builder.WriteString("		pairs = append(pairs, NewStringCodePair(100, \"AcDbSymbolTableRecord\"))\n")
		builder.WriteString("		pairs = append(pairs, item.codePairs(version)...)\n")
		builder.WriteString("	}\n")
		builder.WriteString("	pairs = append(pairs, NewStringCodePair(0, \"ENDTAB\"))\n")
		builder.WriteString("	return\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// codePairs
		builder.WriteString(fmt.Sprintf("func (this *%s) codePairs(version AcadVersion) (pairs []CodePair) {\n", tableItem.Name))
		builder.WriteString("	if version >= R13 {\n")
		builder.WriteString("		pairs = append(pairs, NewStringCodePair(100, \"AcDbSymbolTableRecord\"))\n")
		builder.WriteString(fmt.Sprintf("		pairs = append(pairs, NewStringCodePair(100, \"%s\"))\n", tableItem.ClassName))
		builder.WriteString("	}\n")
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
	builder.WriteString("\n")
	builder.WriteString("	// swallow until 0/<item>\n")
	builder.WriteString("	for error == nil && nextPair.Code != 0 {\n")
	builder.WriteString("		nextPair, error = reader.readCodePair()\n")
	builder.WriteString("	}\n")
	builder.WriteString("\n")
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
		indent := ""
		if len(table.MinVersion) > 0 {
			builder.WriteString(fmt.Sprintf("	if version >= %s {\n", table.MinVersion))
			indent = "	"
		}
		handleFieldName := getHandleFieldName(&table)
		builder.WriteString(fmt.Sprintf("	%spairs = append(pairs, tablePairs%s(drawing.%s, drawing.%s, version)...)\n", indent, table.Collection, handleFieldName, table.Collection))
		if len(table.MinVersion) > 0 {
			builder.WriteString("	}\n")
		}
	}
	builder.WriteString("	return\n")
	builder.WriteString("}\n")

	// assign handles
	builder.WriteString("func assignTableHandles(drawing *Drawing, nextHandle Handle) Handle {\n")
	for _, table := range tables {
		handleFieldName := getHandleFieldName(&table)
		builder.WriteString(fmt.Sprintf("	drawing.%s = nextHandle\n", handleFieldName))
		builder.WriteString("	nextHandle++\n")
		builder.WriteString(fmt.Sprintf("	for i := range drawing.%s {\n", table.Collection))
		builder.WriteString(fmt.Sprintf("		item := &drawing.%s[i]\n", table.Collection))
		builder.WriteString("		if (*item).Handle() == 0 {\n")
		builder.WriteString("			(*item).SetHandle(nextHandle)\n")
		builder.WriteString("			nextHandle++\n")
		builder.WriteString("		}\n")
		builder.WriteString("	}\n")
	}
	builder.WriteString("	return nextHandle\n")
	builder.WriteString("}\n")
	builder.WriteString("\n")

	writeFile("tables.generated.go", builder)
}

func getHandleFieldName(table *xmlTable) string {
	return fmt.Sprintf("%s%sTableHandle", strings.ToLower(table.Collection[:1]), table.Collection[1:len(table.Collection)-1])
}

func readTables(reader io.Reader) ([]xmlTable, error) {
	var tables xmlTables
	if err := xml.NewDecoder(reader).Decode(&tables); err != nil {
		return nil, err
	}

	return tables.Tables, nil
}
