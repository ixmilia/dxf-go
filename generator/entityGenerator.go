package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type xmlSpecification struct {
	XMLName    xml.Name       `xml:"Specification"`
	Entities   []xmlEntity    `xml:"Entity"`
	Interfaces []xmlInterface `xml:"Interface"`
}

type xmlInterface struct {
	XMLName    xml.Name                `xml:"Interface"`
	Name       string                  `xml:"Name,attr"`
	Methods    []xmlMethod             `xml:"Method"`
	Fields     []xmlField              `xml:"Field"`
	WriteOrder xmlWriteOrderCollection `xml:"WriteOrder"`
}

type xmlMethod struct {
	XMLName   xml.Name `xml:"Method"`
	Signature string   `xml:"Signature,attr"`
}

type xmlEntity struct {
	XMLName             xml.Name                `xml:"Entity"`
	Name                string                  `xml:"Name,attr"`
	SubclassMarker      string                  `xml:"SubclassMarker,attr"`
	TypeString          string                  `xml:"TypeString,attr"`
	MinVersion          string                  `xml:"MinVersion,attr"`
	MaxVersion          string                  `xml:"MaxVersion,attr"`
	ConstructorFunction string                  `xml:"ConstructorFunction,attr"`
	GenerateReader      bool                    `xml:"GenerateReader,attr"`
	GenerateWriter      bool                    `xml:"GenerateWriter,attr"`
	ImplementInterfaces string                  `xml:"ImplementInterfaces,attr"`
	Tag                 string                  `xml:"Tag,attr"`
	Fields              []xmlField              `xml:"Field"`
	WriteOrder          xmlWriteOrderCollection `xml:"WriteOrder"`
	Interfaces          []string
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
	Field          string                   `xml:"Field,attr"`
	Code           int                      `xml:"Code,attr"`
	Value          string                   `xml:"Value,attr"`
	WriteCondition string                   `xml:"WriteCondition,attr"`
	WriteConverter string                   `xml:"WriteConverter,attr"`
	MinVersion     string                   `xml:"MinVersion,attr"`
	MaxVersion     string                   `xml:"MaxVersion,attr"`
	Directives     []xmlWriteOrderDirective `xml:",any"`
}

type getNamedField func(string) xmlField

func generateEntities() {
	specPath := "spec/EntitySpec.xml"
	file, err := os.Open(specPath)
	check(err)

	defer file.Close()

	spec, err := readSpecification(file)
	check(err)

	var builder strings.Builder
	builder.WriteString("// Code generated at build time; DO NOT EDIT.\n")
	builder.WriteString("\n")
	builder.WriteString("package dxf\n")
	builder.WriteString("\n")
	builder.WriteString("import (\n")
	builder.WriteString("	\"errors\"\n")
	builder.WriteString("	\"fmt\"\n")
	builder.WriteString("	\"math\"\n")
	builder.WriteString(")\n")
	builder.WriteString("\n")

	interfaces := make(map[string]xmlInterface)
	for _, inf := range spec.Interfaces {
		interfaces[inf.Name] = inf
	}

	// output interfaces
	for _, inf := range spec.Interfaces {
		// base interface
		builder.WriteString(fmt.Sprintf("type %s interface {\n", inf.Name))
		for _, method := range inf.Methods {
			builder.WriteString(fmt.Sprintf("	%s\n", method.Signature))
		}
		for _, field := range inf.Fields {
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
		builder.WriteString(fmt.Sprintf("func tryApplyCodePairFor%s(this %s, codePair CodePair) bool {\n", inf.Name, inf.Name))
		builder.WriteString("	switch codePair.Code {\n")
		for _, field := range inf.Fields {
			readField(&builder, field, true)
		}
		builder.WriteString("	default:\n")
		builder.WriteString("		return false\n")
		builder.WriteString("	}\n")
		builder.WriteString("	return true\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// code pair builder
		builder.WriteString(fmt.Sprintf("func codePairsFor%s(this %s, version AcadVersion) (pairs []CodePair) {\n", inf.Name, inf.Name))
		if len(inf.WriteOrder.Directives) > 0 {
			for _, directive := range inf.WriteOrder.Directives {
				writeDirective(&builder, directive, inf.getNamedField, true, "")
			}
		} else {
			for _, field := range inf.Fields {
				writeField(&builder, field, true, "")
			}
		}
		builder.WriteString("	return\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")
	}

	// for each entity
	for _, entity := range spec.Entities {
		// declaration
		builder.WriteString(fmt.Sprintf("type %s struct {\n", entity.Name))

		// backing interface
		for _, infName := range entity.Interfaces {
			inf := interfaces[infName]
			builder.WriteString(fmt.Sprintf("	// fields for %s interface\n", inf.Name))
			for _, field := range inf.Fields {
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
		}

		// specific fields
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
		for _, infName := range entity.Interfaces {
			inf := interfaces[infName]
			for _, field := range inf.Fields {
				backingField := strings.ToLower(field.Name[0:1]) + field.Name[1:]
				builder.WriteString(fmt.Sprintf("		%s: %s,\n", backingField, field.DefaultValue))
			}
		}
		for _, field := range entity.Fields {
			builder.WriteString(fmt.Sprintf("		%s: %s,\n", field.Name, field.DefaultValue))
		}
		builder.WriteString("	}\n")
		builder.WriteString("}\n")
		builder.WriteString("\n")

		// base interface getter/setter
		for _, infName := range entity.Interfaces {
			inf := interfaces[infName]
			for _, field := range inf.Fields {
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
			for _, field := range entity.Fields {
				readField(&builder, field, false)
			}
			builder.WriteString("	default:\n")
			builder.WriteString("		appliedCodePair := false\n")
			for i := len(entity.Interfaces) - 1; i >= 0; i-- {
				infName := entity.Interfaces[i]
				builder.WriteString("		if !appliedCodePair {\n")
				builder.WriteString(fmt.Sprintf("			appliedCodePair = tryApplyCodePairFor%s(this, codePair)\n", infName))
				builder.WriteString("		}\n")
			}
			builder.WriteString("	}\n")
			builder.WriteString("}\n")
			builder.WriteString("\n")
		}

		// writer
		if entity.GenerateWriter {
			builder.WriteString(fmt.Sprintf("func (this *%s) codePairs(version AcadVersion) (pairs []CodePair) {\n", entity.Name))
			builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(0, \"%s\"))\n", strings.Split(entity.TypeString, ",")[0]))
			for _, infName := range entity.Interfaces {
				inf := interfaces[infName]
				builder.WriteString(fmt.Sprintf("	pairs = append(pairs, codePairsFor%s(this, version)...)\n", inf.Name))
			}
			if len(entity.WriteOrder.Directives) > 0 {
				for _, directive := range entity.WriteOrder.Directives {
					writeDirective(&builder, directive, entity.getNamedField, false, "")
				}
			} else {
				builder.WriteString(fmt.Sprintf("	pairs = append(pairs, NewStringCodePair(100, \"%s\"))\n", entity.SubclassMarker))
				for _, field := range entity.Fields {
					writeField(&builder, field, false, "")
				}
			}
			builder.WriteString("	return\n")
			builder.WriteString("}\n")
			builder.WriteString("\n")
		}
	}

	// dimension creator
	builder.WriteString("func createAndPopulateDimension(temp *dimensionHelper) (dimension Entity, error error) {\n")
	builder.WriteString("	switch temp.DimensionType() {\n")
	for _, dim := range spec.Entities {
		if dim.implementsInterface("Dimension") && dim.Name != "dimensionHelper" {
			builder.WriteString(fmt.Sprintf("	case DimensionType%s:\n", dim.Tag))
			builder.WriteString(fmt.Sprintf("		dimension = New%s()\n", dim.Name))
		}
	}
	builder.WriteString("	default:\n")
	builder.WriteString("		error = errors.New(fmt.Sprintf(\"Unsupported dimension type %s\", temp.DimensionType()))\n")
	builder.WriteString("		return\n")
	builder.WriteString("	}\n")
	builder.WriteString("\n")
	builder.WriteString("	for _, pair := range temp.collectedPairs {\n")
	builder.WriteString("		dimension.tryApplyCodePair(pair)\n")
	builder.WriteString("	}\n")
	builder.WriteString("\n")
	builder.WriteString("	return\n")
	builder.WriteString("}\n")
	builder.WriteString("\n")

	// entity creator
	builder.WriteString("func createEntity(entityType string) (entity Entity, ok bool) {\n")
	builder.WriteString("	ok = true\n")
	builder.WriteString("	switch entityType {\n")
	seenTypeStrings := make(map[string]bool)
	for _, entity := range spec.Entities {
		if seenTypeStrings[entity.TypeString] {
			continue
		}
		seenTypeStrings[entity.TypeString] = true
		typeStrings := strings.Split(entity.TypeString, ",")
		for i := 0; i < len(typeStrings); i++ {
			typeStrings[i] = "\"" + typeStrings[i] + "\""
		}
		constructorFunction := fmt.Sprintf("New%s()", entity.Name)
		if len(entity.ConstructorFunction) > 0 {
			constructorFunction = entity.ConstructorFunction
		}
		builder.WriteString(fmt.Sprintf("	case %s:\n", strings.Join(typeStrings, ", ")))
		builder.WriteString(fmt.Sprintf("		entity = %s\n", constructorFunction))
	}
	builder.WriteString("	default:\n")
	builder.WriteString("		ok = false\n")
	builder.WriteString("	}\n")
	builder.WriteString("	return\n")
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

func writeDirective(builder *strings.Builder, directive xmlWriteOrderDirective, getNamedField getNamedField, asFunction bool, indent string) {
	switch directive.XMLName.Local {
	case "Foreach":
		builder.WriteString(fmt.Sprintf("%s	for _, item := range this.%s {\n", indent, directive.Field))
		for _, d := range directive.Directives {
			writeDirective(builder, d, getNamedField, asFunction, indent+"\t")
		}
		builder.WriteString(indent + "	}\n")
	case "WriteExtensionData":
		// TODO:
	case "WriteField":
		field := getNamedField(directive.Field)
		writeField(builder, field, asFunction, indent)
	case "WriteSpecificValue":
		predicates := directivePredicates(directive)
		if len(predicates) > 0 {
			builder.WriteString(fmt.Sprintf("	if %s {\n", strings.Join(predicates, " && ")))
			indent += "\t"
		}
		builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, New%sCodePair(%d, %s))\n", indent, codeTypeName(directive.Code), directive.Code, directive.Value))
		if len(predicates) > 0 {
			builder.WriteString("	}\n")
		}
	default:
		panic(fmt.Sprintf("Unsupported write directive '%s'.", directive.XMLName.Local))
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

func readField(builder *strings.Builder, field xmlField, asInterface bool) {
	if field.Code < 0 {
		// specially handled, just needs to exist
		return
	}

	suffix := ""
	if asInterface {
		suffix = "()"
	}

	if len(field.CodeOverrides) > 0 {
		codeOverrides := strings.Split(field.CodeOverrides, ",")
		for i, codeString := range codeOverrides {
			code, err := strconv.Atoi(strings.TrimSpace(codeString))
			check(err)
			component := 'X' + i
			builder.WriteString(fmt.Sprintf("	case %d:\n", code))
			pairValue := "codePair.Value.(DoubleCodePairValue).Value"
			if asInterface {
				builder.WriteString(fmt.Sprintf("		temp := this.%s()\n", field.Name))
				builder.WriteString(fmt.Sprintf("		temp.%c = %s\n", component, pairValue))
				builder.WriteString(fmt.Sprintf("		this.Set%s(temp)\n", field.Name))
			} else {
				builder.WriteString(fmt.Sprintf("		this.%s.%c = %s\n", field.Name, component, pairValue))
			}
		}
	} else {
		builder.WriteString(fmt.Sprintf("	case %d:\n", field.Code))
		readValue := fmt.Sprintf("codePair.Value.(%sCodePairValue).Value", codeTypeName(field.Code))
		if len(field.ReadConverter) > 0 {
			readValue = strings.Replace(field.ReadConverter, "%v", readValue, -1)
		}
		if field.AllowMultiples {
			readValue = fmt.Sprintf("append(this.%s%s, %s)", field.Name, suffix, readValue)
		}

		if asInterface {
			builder.WriteString(fmt.Sprintf("		this.Set%s(%s)\n", field.Name, readValue))
		} else {
			builder.WriteString(fmt.Sprintf("		this.%s = %s\n", field.Name, readValue))
		}
	}
}

func writeField(builder *strings.Builder, field xmlField, asInterface bool, indent string) {
	if field.Code < 0 {
		// specially handled, just needs to exist
		return
	}

	predicates := fieldPredicates(field, asInterface)
	if len(predicates) > 0 {
		indent += "\t"
		builder.WriteString(fmt.Sprintf("%sif %s {\n", indent, strings.Join(predicates, " && ")))
	}

	suffix := ""
	if asInterface {
		suffix = "()"
	}

	if len(field.CodeOverrides) > 0 {
		codeOverrides := strings.Split(field.CodeOverrides, ",")
		for i, codeString := range codeOverrides {
			code, err := strconv.Atoi(strings.TrimSpace(codeString))
			check(err)
			component := 'X' + i
			builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, NewDoubleCodePair(%d, this.%s%s.%c))\n", indent, code, field.Name, suffix, component))
		}
	} else {
		if field.AllowMultiples {
			builder.WriteString(fmt.Sprintf("%s	for _, val := range this.%s%s {\n", indent, field.Name, suffix))
			value := "val"
			if len(field.WriteConverter) > 0 {
				value = strings.Replace(field.WriteConverter, "%v", value, -1)
			}
			builder.WriteString(fmt.Sprintf("%s		pairs = append(pairs, New%sCodePair(%d, %s))\n", indent, codeTypeName(field.Code), field.Code, value))
			builder.WriteString(fmt.Sprintf("%s	}\n", indent))
		} else {
			value := fmt.Sprintf("this.%s%s", field.Name, suffix)
			if len(field.WriteConverter) > 0 {
				value = strings.Replace(field.WriteConverter, "%v", value, -1)
			}
			builder.WriteString(fmt.Sprintf("%s	pairs = append(pairs, New%sCodePair(%d, %s))\n", indent, codeTypeName(field.Code), field.Code, value))
		}
	}
	if len(predicates) > 0 {
		builder.WriteString(indent + "}\n")
	}
}

func fieldPredicates(field xmlField, asInterface bool) (predicates []string) {
	if len(field.MinVersion) > 0 {
		predicates = append(predicates, fmt.Sprintf("version >= %s", field.MinVersion))
	}
	if len(field.MaxVersion) > 0 {
		predicates = append(predicates, fmt.Sprintf("version <= %s", field.MaxVersion))
	}
	if field.DisableWritingDefault {
		suffix := ""
		if asInterface {
			suffix = "()"
		}
		predicates = append(predicates, fmt.Sprintf("this.%s%s != %s", field.Name, suffix, field.DefaultValue))
	}

	return
}

func (entity xmlEntity) implementsInterface(interfaceName string) bool {
	for _, inf := range entity.Interfaces {
		if inf == interfaceName {
			return true
		}
	}

	return false
}

func (entity xmlEntity) getNamedField(name string) xmlField {
	for _, field := range entity.Fields {
		if field.Name == name {
			return field
		}
	}

	panic(fmt.Sprintf("Unable to find field %s.%s", entity.Name, name))
}

func (inf xmlInterface) getNamedField(name string) xmlField {
	for _, field := range inf.Fields {
		if field.Name == name {
			return field
		}
	}

	panic(fmt.Sprintf("Unable to find field %s.%s", inf.Name, name))
}

func (entity *xmlEntity) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type tempXMLEntity xmlEntity

	// set non-standard defaults
	item := tempXMLEntity{
		GenerateReader:      true,
		GenerateWriter:      true,
		ImplementInterfaces: "Entity",
	}
	err := d.DecodeElement(&item, &start)
	if err != nil {
		return err
	}
	item.Interfaces = strings.Split(item.ImplementInterfaces, ",")
	*entity = (xmlEntity)(item)
	return nil
}

func readSpecification(reader io.Reader) (spec xmlSpecification, error error) {
	error = xml.NewDecoder(reader).Decode(&spec)
	return
}
