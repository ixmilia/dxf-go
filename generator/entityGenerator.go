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
	MinVersion     string                  `xml:"MinVersion,attr"`
	MaxVersion     string                  `xml:"MaxVersion,attr"`
	GenerateReader bool                    `xml:"GenerateReader,attr"`
	Fields         []xmlField              `xml:"Field"`
	WriteOrder     xmlWriteOrderCollection `xml:"WriteOrder"`
}

type xmlField struct {
	XMLName               xml.Name  `xml:"Field"`
	Name                  string    `xml:"Name,attr"`
	Code                  int       `xml:"Code,attr"`
	CodeOverrides         string    `xml:"CodeOverrides,attr"`
	Type                  string    `xml:"Type,attr"`
	DefaultValue          string    `xml:"DefaultValue,attr"`
	ReadConverter         string    `xml:"ReadConverter,attr"`
	WriteConverter        string    `xml:"WriteConverter,attr"`
	DisableWritingDefault bool      `xml:"DisableWritingDefault,attr"`
	AllowMultiples        bool      `xml:"AllowMultiples,attr"`
	MinVersion            string    `xml:"MinVersion,attr"`
	MaxVersion            string    `xml:"MaxVersion,attr"`
	Comment               string    `xml:"Comment,attr"`
	Flags                 []xmlFlag `xml:"Flag"`
}

type xmlFlag struct {
	XMLName xml.Name `xml:"Flag"`
	Name    string   `xml:"Name,attr"`
	Mask    int      `xml:"Mask,attr"`
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
	MinVersion     string `xml:"MinVersion,attr"`
	MaxVersion     string `xml:"MaxVersion,attr"`
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
	builder.WriteString("	minVersion() (version AcadVersion)\n")
	builder.WriteString("	maxVersion() (version AcadVersion)\n")
	builder.WriteString("	codePairs(version AcadVersion) (pairs []CodePair)\n")
	builder.WriteString("	tryApplyCodePair(codePair CodePair)\n")
	for _, field := range baseEntity.Fields {
		fieldType := field.Type
		if field.AllowMultiples {
			fieldType = "[]" + fieldType
		}
		builder.WriteString(fmt.Sprintf("	%s() %s\n", field.Name, fieldType))       // getter
		builder.WriteString(fmt.Sprintf("	Set%s(val %s)\n", field.Name, fieldType)) // setter
	}
	builder.WriteString("}\n")
	builder.WriteString("\n")

	// base reader
	builder.WriteString("func tryApplyBaseCodePair(entity Entity, codePair CodePair) {\n")
	builder.WriteString("	switch codePair.Code {\n")
	for _, field := range baseEntity.Fields {
		builder.WriteString(fmt.Sprintf("	case %d:\n", field.Code))
		readValue := fmt.Sprintf("codePair.Value.(%sCodePairValue).Value", codeTypeName(field.Code))
		if len(field.ReadConverter) > 0 {
			readValue = strings.Replace(field.ReadConverter, "%v", readValue, -1)
		}
		if field.AllowMultiples {
			readValue = fmt.Sprintf("append(entity.%s(), %s)", field.Name, readValue)
		}
		builder.WriteString(fmt.Sprintf("		entity.Set%s(%s)\n", field.Name, readValue))
	}
	builder.WriteString("	}\n")
	builder.WriteString("}\n")
	builder.WriteString("\n")

	// for each entity
	for _, entity := range entities {
		if entity.Name == "Entity" {
			continue
		}

		// declaration
		builder.WriteString(fmt.Sprintf("type %s struct {\n", entity.Name))
		// backing fields
		for _, field := range baseEntity.Fields {
			comment := ""
			if len(field.Comment) > 0 {
				comment = fmt.Sprintf(" // %s", field.Comment)
			}
			backingField := strings.ToLower(field.Name[0:1]) + field.Name[1:]
			fieldType := field.Type
			if field.AllowMultiples {
				fieldType = "[]" + fieldType
			}

			builder.WriteString(fmt.Sprintf("	%s %s%s\n", backingField, fieldType, comment))
		}
		for _, field := range entity.Fields {
			comment := ""
			if len(field.Comment) > 0 {
				comment = fmt.Sprintf(" // %s", field.Comment)
			}
			fieldType := field.Type
			if field.AllowMultiples {
				fieldType = "[]" + fieldType
			}
			builder.WriteString(fmt.Sprintf("	%s %s%s\n", field.Name, fieldType, comment))
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

		// base entity getter/setter
		for _, field := range baseEntity.Fields {
			fieldType := field.Type
			if field.AllowMultiples {
				fieldType = "[]" + fieldType
			}

			// getter
			builder.WriteString(fmt.Sprintf("func (this *%s) %s() %s {\n", entity.Name, field.Name, fieldType))
			backingField := strings.ToLower(field.Name[0:1]) + field.Name[1:]
			builder.WriteString(fmt.Sprintf("	return this.%s\n", backingField))
			builder.WriteString("}\n")
			builder.WriteString("\n")

			// setter
			builder.WriteString(fmt.Sprintf("func (this *%s) Set%s(val %s) {\n", entity.Name, field.Name, fieldType))
			builder.WriteString(fmt.Sprintf("	this.%s = val\n", backingField))
			builder.WriteString("}\n")
			builder.WriteString("\n")
		}

		// flags
		for _, field := range entity.Fields {
			for _, flag := range field.Flags {
				comment := generateComment(fmt.Sprintf("%s status flag.", flag.Name), field.MinVersion, field.MaxVersion)

				// getter
				builder.WriteString(fmt.Sprintf("// %s\n", comment))
				builder.WriteString(fmt.Sprintf("func (this *%s) %s() bool {\n", entity.Name, flag.Name))
				builder.WriteString(fmt.Sprintf("	return this.%s & %d != 0\n", field.Name, flag.Mask))
				builder.WriteString("}\n")
				builder.WriteString("\n")

				// setter
				builder.WriteString(fmt.Sprintf("// %s\n", comment))
				builder.WriteString(fmt.Sprintf("func (this *%s) Set%s(val bool) {\n", entity.Name, flag.Name))
				builder.WriteString("	if val {\n")
				builder.WriteString(fmt.Sprintf("		this.%s = this.%s | %d\n", field.Name, field.Name, flag.Mask))
				builder.WriteString("	} else {\n")
				builder.WriteString(fmt.Sprintf("		this.%s = this.%s & ^%d\n", field.Name, field.Name, flag.Mask))
				builder.WriteString("	}\n")
				builder.WriteString("}\n")
				builder.WriteString("\n")
			}
		}

		collectionHelpers(&builder, entity, entity.Name)

		// typeString()
		builder.WriteString(fmt.Sprintf("func (this *%s) typeString() string {\n", entity.Name))
		builder.WriteString(fmt.Sprintf("	return \"%s\"\n", strings.Split(entity.TypeString, ",")[0]))
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// minVersion()
		minVersion := entity.MinVersion
		if len(minVersion) == 0 {
			minVersion = "Version1_0" // TODO: pull this from acadVersion.go?
		}
		builder.WriteString(fmt.Sprintf("func (this *%s) minVersion() (version AcadVersion) {\n", entity.Name))
		builder.WriteString(fmt.Sprintf("	return %s\n", minVersion))
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// maxVersion()
		maxVersion := entity.MaxVersion
		if len(maxVersion) == 0 {
			maxVersion = "R2018" // TODO: pull this from acadVersion.go?
		}
		builder.WriteString(fmt.Sprintf("func (this *%s) maxVersion() (version AcadVersion) {\n", entity.Name))
		builder.WriteString(fmt.Sprintf("	return %s\n", maxVersion))
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// reader
		if entity.GenerateReader {
			builder.WriteString(fmt.Sprintf("func (this *%s) tryApplyCodePair(codePair CodePair) {\n", entity.Name))
			builder.WriteString("	switch codePair.Code {\n")
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
					if field.AllowMultiples {
						readValue = fmt.Sprintf("append(this.%s, %s)", field.Name, readValue)
					}

					builder.WriteString(fmt.Sprintf("		this.%s = %s\n", field.Name, readValue))
				}

			}
			builder.WriteString("	default:\n")
			builder.WriteString("		tryApplyBaseCodePair(this, codePair)\n")
			builder.WriteString("	}\n")
			builder.WriteString("}\n")
			builder.WriteString("\n")
		}

		// writer
		builder.WriteString(fmt.Sprintf("func (this *%s) codePairs(version AcadVersion) (pairs []CodePair) {\n", entity.Name))
		builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(0, \"%s\"))\n", strings.Split(entity.TypeString, ",")[0]))
		for _, directive := range baseEntity.WriteOrder.Directives {
			writeDirective(&builder, baseEntity, directive, true)
		}
		if len(entity.WriteOrder.Directives) > 0 {
			for _, directive := range entity.WriteOrder.Directives {
				writeDirective(&builder, entity, directive, false)
			}
		} else {
			builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(100, \"%s\"))\n", entity.SubclassMarker))
			for _, field := range entity.Fields {
				writeField(&builder, entity, field, false)
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

func collectionHelpers(builder *strings.Builder, entity xmlEntity, entityName string) {
	for _, field := range entity.Fields {
		if field.AllowMultiples {
			// add
			builder.WriteString(fmt.Sprintf("func (this *%s) Add%s(val %s) {\n", entityName, field.Name, field.Type))
			builder.WriteString(fmt.Sprintf("	this.%s = append(this.%s, val)\n", field.Name, field.Name))
			builder.WriteString("}\n")
			builder.WriteString("\n")

			// clear
			builder.WriteString(fmt.Sprintf("func (this *%s) Clear%s() {\n", entityName, field.Name))
			builder.WriteString(fmt.Sprintf("	this.%s = []%s{}\n", field.Name, field.Type))
			builder.WriteString("}\n")
			builder.WriteString("\n")
		}
	}
}

func writeDirective(builder *strings.Builder, entity xmlEntity, directive xmlWriteOrderDirective, asFunction bool) {
	switch directive.XMLName.Local {
	case "WriteExtensionData":
		// TODO:
	case "WriteField":
		field := entity.getNamedField(directive.Field)
		writeField(builder, entity, field, asFunction)
	case "WriteSpecificValue":
		predicates := directivePredicates(directive)
		indention := ""
		if len(predicates) > 0 {
			builder.WriteString(fmt.Sprintf("	if %s {\n", strings.Join(predicates, " && ")))
			indention = "	"
		}
		builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, New%sCodePair(%d, %s))\n", indention, codeTypeName(directive.Code), directive.Code, directive.Value))
		if len(predicates) > 0 {
			builder.WriteString("	}\n")
		}
	default:
		panic(fmt.Sprintf("Unsupported write directive '%s' specified for entity %s", directive.XMLName.Local, entity.Name))
	}
}

func directivePredicates(directive xmlWriteOrderDirective) []string {
	predicates := []string{}
	if len(directive.MinVersion) > 0 {
		predicates = append(predicates, fmt.Sprintf("version >= %s", directive.MinVersion))
	}
	if len(directive.MaxVersion) > 0 {
		predicates = append(predicates, fmt.Sprintf("version <= %s", directive.MaxVersion))
	}
	if len(directive.WriteCondition) > 0 {
		predicates = append(predicates, directive.WriteCondition)
	}
	return predicates
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

	if len(field.CodeOverrides) > 0 {
		codeOverrides := strings.Split(field.CodeOverrides, ",")
		for i, codeString := range codeOverrides {
			code, err := strconv.Atoi(strings.TrimSpace(codeString))
			check(err)
			component := 'X' + i
			builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, NewDoubleCodePair(%d, this.%s.%c))\n", indention, code, field.Name, component))
		}
	} else {
		if field.AllowMultiples {
			builder.WriteString(fmt.Sprintf("%s	for _, val := range this.%s%s {\n", indention, field.Name, suffix))
			value := "val"
			if len(field.WriteConverter) > 0 {
				value = strings.Replace(field.WriteConverter, "%v", value, -1)
			}
			builder.WriteString(fmt.Sprintf("%s		pairs = append(pairs, New%sCodePair(%d, %s))\n", indention, codeTypeName(field.Code), field.Code, value))
			builder.WriteString(fmt.Sprintf("%s	}\n", indention))
		} else {
			value := fmt.Sprintf("this.%s%s", field.Name, suffix)
			if len(field.WriteConverter) > 0 {
				value = strings.Replace(field.WriteConverter, "%v", value, -1)
			}
			builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, New%sCodePair(%d, %s))\n", indention, codeTypeName(field.Code), field.Code, value))
		}
	}
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

func (entity *xmlEntity) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type tempXMLEntity xmlEntity

	// set non-standard defaults
	item := tempXMLEntity{
		GenerateReader: true,
	}
	err := d.DecodeElement(&item, &start)
	if err != nil {
		return err
	}
	*entity = (xmlEntity)(item)
	return nil
}

func readEntities(reader io.Reader) ([]xmlEntity, error) {
	var entities xmlEntities
	if err := xml.NewDecoder(reader).Decode(&entities); err != nil {
		return nil, err
	}

	return entities.Entities, nil
}
