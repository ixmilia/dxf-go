package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type xmlEntities struct {
	XMLName  xml.Name    `xml:"Entities"`
	Entities []xmlEntity `xml:"Entity"`
}

type xmlEntity struct {
	XMLName        xml.Name                `xml:"Entity"`
	Name           string                  `xml:"Name,attr"`
	SubclassMarker string                  `xml:"SubclassMarker,attr"`
	TypeString     string                  `xml:"TypeString,attr"`
	Fields         []xmlField              `xml:"Field"`
	WriteOrder     xmlWriteOrderCollection `xml:"WriteOrder"`
}

type xmlField struct {
	XMLName               xml.Name `xml:"Field"`
	Name                  string   `xml:"Name,attr"`
	Code                  int      `xml:"Code,attr"`
	CodeOverrides         string   `xml:"CodeOverrides,attr"`
	Type                  string   `xml:"Type,attr"`
	DefaultValue          string   `xml:"DefaultValue,attr"`
	ReadConverter         string   `xml:"ReadConverter,attr"`
	WriteConverter        string   `xml:"WriteConverter,attr"`
	DisableWritingDefault bool     `xml:"DisableWritingDefault,attr"`
	MinVersion            string   `xml:"MinVersion,attr"`
	MaxVersion            string   `xml:"MaxVersion,attr"`
	Comment               string   `xml:"Comment,attr"`
}

type xmlWriteOrderCollection struct {
	XMLName    xml.Name                 `xml:"WriteOrder"`
	Directives []xmlWriteOrderDirective `xml:",any"`
}

type xmlWriteOrderDirective struct {
	XMLName        xml.Name
	Field          string `xml:"Field,attr"`
	Code           int    `xml:"Code,attr"`
	Value          string `xml:"Value,attr"`
	WriteCondition string `xml:"WriteCondition,attr"`
	WriteConverter string `xml:"WriteConverter,attr"`
}

func generateEntities() {
	specPath := "spec/EntitySpec.xml"
	file, err := os.Open(specPath)
	check(err)

	defer file.Close()

	entities, err := readEntities(file)
	check(err)

	var builder strings.Builder
	builder.WriteString("// Code generated at build time; DO NOT EDIT.\n")
	builder.WriteString("\n")
	builder.WriteString("package dxf\n")
	builder.WriteString("\n")

	var baseEntity xmlEntity
	foundBaseEntity := false
	for _, entity := range entities {
		if entity.Name == "Entity" {
			baseEntity = entity
			foundBaseEntity = true
			break
		}
	}

	if !foundBaseEntity {
		panic("unable to find base entity")
	}

	// base interface
	builder.WriteString("type Entity interface {\n")
	builder.WriteString("	typeString() (typeString string)\n")
	builder.WriteString("	codePairs(version AcadVersion) (pairs []CodePair)\n")
	builder.WriteString("	tryApplyCodePair(codePair CodePair)\n")
	for _, field := range baseEntity.Fields {
		builder.WriteString(fmt.Sprintf("	%s() %s\n", field.Name, field.Type))       // getter
		builder.WriteString(fmt.Sprintf("	Set%s(val %s)\n", field.Name, field.Type)) // setter
	}
	builder.WriteString("}\n")
	builder.WriteString("\n")

	for _, entity := range entities {
		if entity.Name == "Entity" {
			continue
		}

		// declaration
		builder.WriteString(fmt.Sprintf("type %s struct {\n", entity.Name))
		// backing fields
		for _, field := range baseEntity.Fields {
			backingField := strings.ToLower(field.Name[0:1]) + field.Name[1:]
			builder.WriteString(fmt.Sprintf("	%s %s\n", backingField, field.Type))
		}
		for _, field := range entity.Fields {
			// TODO: allow multiples
			builder.WriteString(fmt.Sprintf("	%s %s\n", field.Name, field.Type))
		}
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// constructor
		builder.WriteString(fmt.Sprintf("func New%s() *%s {\n", entity.Name, entity.Name))
		builder.WriteString(fmt.Sprintf("	return &%s{\n", entity.Name))
		for _, field := range baseEntity.Fields {
			backingField := strings.ToLower(field.Name[0:1]) + field.Name[1:]
			builder.WriteString(fmt.Sprintf("		%s: %s,\n", backingField, field.DefaultValue))
		}
		for _, field := range entity.Fields {
			builder.WriteString(fmt.Sprintf("		%s: %s,\n", field.Name, field.DefaultValue))
		}
		builder.WriteString("	}\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// getter/setter
		for _, field := range baseEntity.Fields {
			// getter
			builder.WriteString(fmt.Sprintf("func (this *%s) %s() %s {\n", entity.Name, field.Name, field.Type))
			backingField := strings.ToLower(field.Name[0:1]) + field.Name[1:]
			builder.WriteString(fmt.Sprintf("	return this.%s\n", backingField))
			builder.WriteString("}\n")
			builder.WriteString("\n")

			// setter
			builder.WriteString(fmt.Sprintf("func (this *%s) Set%s(val %s) {\n", entity.Name, field.Name, field.Type))
			builder.WriteString(fmt.Sprintf("	this.%s = val\n", backingField))
			builder.WriteString("}\n")
			builder.WriteString("\n")
		}

		builder.WriteString(fmt.Sprintf("func (this *%s) typeString() string {\n", entity.Name))
		builder.WriteString(fmt.Sprintf("	return \"%s\"\n", strings.Split(entity.TypeString, ",")[0]))
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// reader
		builder.WriteString(fmt.Sprintf("func (this *%s) tryApplyCodePair(codePair CodePair) {\n", entity.Name))
		builder.WriteString("	switch codePair.Code {\n")
		builder.WriteString("	// base Entity values\n")
		for _, field := range baseEntity.Fields {
			builder.WriteString(fmt.Sprintf("	case %d:\n", field.Code))
			readValue := fmt.Sprintf("codePair.Value.(%sCodePairValue).Value", codeTypeName(field.Code))
			if len(field.ReadConverter) > 0 {
				readValue = strings.Replace(field.ReadConverter, "%v", readValue, -1)
			}
			builder.WriteString(fmt.Sprintf("		this.Set%s(%s)\n", field.Name, readValue))
		}
		builder.WriteString("\n")
		builder.WriteString("	// entity specific values\n")
		for _, field := range entity.Fields {
			if len(field.CodeOverrides) > 0 {
				codeOverrides := strings.Split(field.CodeOverrides, ",")
				for i, codeString := range codeOverrides {
					code, err := strconv.Atoi(strings.TrimSpace(codeString))
					check(err)
					component := 'X' + i
					builder.WriteString(fmt.Sprintf("	case %d:\n", code))
					builder.WriteString(fmt.Sprintf("		this.%s.%c = codePair.Value.(DoubleCodePairValue).Value\n", field.Name, component))
				}
			} else {
				builder.WriteString(fmt.Sprintf("	case %d:\n", field.Code))
				readValue := fmt.Sprintf("codePair.Value.(%sCodePairValue).Value", codeTypeName(field.Code))
				if len(field.ReadConverter) > 0 {
					readValue = strings.Replace(field.ReadConverter, "%v", readValue, -1)
				}
				builder.WriteString(fmt.Sprintf("		this.%s = %s\n", field.Name, readValue))
			}

		}
		builder.WriteString("	}\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// writer
		builder.WriteString(fmt.Sprintf("func (this *%s) codePairs(version AcadVersion) (pairs []CodePair) {\n", entity.Name))
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(0, \"%s\"))\n", strings.Split(entity.TypeString, ",")[0]))
		for _, directive := range baseEntity.WriteOrder.Directives {
			writeDirective(&builder, baseEntity, directive, true)
		}
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(100, \"%s\"))\n", entity.SubclassMarker))
		if len(entity.WriteOrder.Directives) > 0 {
			for _, directive := range entity.WriteOrder.Directives {
				writeDirective(&builder, entity, directive, false)
			}
		} else {
			for _, field := range entity.Fields {
				if len(field.CodeOverrides) > 0 {
					codeOverrides := strings.Split(field.CodeOverrides, ",")
					for i, codeString := range codeOverrides {
						code, err := strconv.Atoi(strings.TrimSpace(codeString))
						check(err)
						component := 'X' + i
						builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewDoubleCodePair(%d, this.%s.%c))\n", code, field.Name, component))
					}
				} else {
					writeField(&builder, entity, field, false)
				}
			}
		}
		builder.WriteString("\n")
		builder.WriteString("	return pairs\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")
	}

	// entity creator
	builder.WriteString("func createEntity(entityType string) (entity Entity, ok bool) {\n")
	builder.WriteString("	ok = true\n")
	builder.WriteString("	switch entityType {\n")
	for _, entity := range entities {
		if entity.Name == "Entity" {
			continue
		}
		typeStrings := strings.Split(entity.TypeString, ",")
		for i := 0; i < len(typeStrings); i++ {
			typeStrings[i] = "\"" + typeStrings[i] + "\""
		}
		builder.WriteString(fmt.Sprintf("	case %s:\n", strings.Join(typeStrings, ", ")))
		builder.WriteString(fmt.Sprintf("		entity = New%s()\n", entity.Name))
	}
	builder.WriteString("	default:\n")
	builder.WriteString("		ok = false\n")
	builder.WriteString("	}\n")
	builder.WriteString("\n")
	builder.WriteString("	return entity, ok\n")
	builder.WriteString("}\n")

	writeFile("entities.generated.go", builder)
}

func writeDirective(builder *strings.Builder, entity xmlEntity, directive xmlWriteOrderDirective, asFunction bool) {
	switch directive.XMLName.Local {
	case "WriteExtensionData":
		// TODO:
	case "WriteField":
		field := entity.getNamedField(directive.Field)
		writeField(builder, entity, field, asFunction)
	case "WriteSpecificValue":
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, New%sCodePair(%d, %s))\n", codeTypeName(directive.Code), directive.Code, directive.Value))
	default:
		panic(fmt.Sprintf("Unsupported write directive '%s' specified for entity %s", directive.XMLName.Local, entity.Name))
	}
}

func writeField(builder *strings.Builder, entity xmlEntity, field xmlField, asFunction bool) {
	predicates := fieldPredicates(field, asFunction)
	indention := ""
	if len(predicates) > 0 {
		indention = "	"
		builder.WriteString(fmt.Sprintf("	if %s {\n", strings.Join(predicates, " && ")))
	}

	suffix := ""
	if asFunction {
		suffix = "()"
	}
	value := fmt.Sprintf("this.%s%s", field.Name, suffix)
	if len(field.WriteConverter) > 0 {
		value = strings.Replace(field.WriteConverter, "%v", value, -1)
	}
	builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, New%sCodePair(%d, %s))\n", indention, codeTypeName(field.Code), field.Code, value))
	if len(predicates) > 0 {
		builder.WriteString("	}\n")
	}
}

func fieldPredicates(field xmlField, asFunction bool) (predicates []string) {
	if len(field.MinVersion) > 0 {
		predicates = append(predicates, fmt.Sprintf("version >= %s", field.MinVersion))
	}
	if len(field.MaxVersion) > 0 {
		predicates = append(predicates, fmt.Sprintf("version <= %s", field.MaxVersion))
	}
	if field.DisableWritingDefault {
		suffix := ""
		if asFunction {
			suffix = "()"
		}
		predicates = append(predicates, fmt.Sprintf("this.%s%s != %s", field.Name, suffix, field.DefaultValue))
	}

	return
}

func (entity xmlEntity) getNamedField(name string) xmlField {
	for _, field := range entity.Fields {
		if field.Name == name {
			return field
		}
	}

	panic(fmt.Sprintf("Unable to find field %s.%s", entity.Name, name))
}

func readEntities(reader io.Reader) ([]xmlEntity, error) {
	var entities xmlEntities
	if err := xml.NewDecoder(reader).Decode(&entities); err != nil {
		return nil, err
	}

	return entities.Entities, nil
}
