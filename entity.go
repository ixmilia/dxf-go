package dxf

import (
	"encoding/hex"
	"errors"
	"strings"
)

func readEntities(np CodePair, reader codePairReader) (entities []Entity, nextPair CodePair, error error) {
	var entity Entity
	var ok bool
	nextPair = np
	for error == nil && !nextPair.isEndSection() {
		entity, nextPair, ok, error = readEntity(nextPair, reader)
		if error != nil {
			return
		} else if ok {
			entities = append(entities, entity)
		}
		// otherwise an unsupported entity was swallowed
	}

	if error != nil {
		return
	}

	entityBuffer := &entityBufferReader{
		entities: entities,
		position: 0,
	}

	entities = collectEntities(entityBuffer)
	return
}

func readEntity(np CodePair, reader codePairReader) (entity Entity, nextPair CodePair, created bool, error error) {
	nextPair = np
	if nextPair.Code != 0 {
		created = false
		error = errors.New("exepcted 0/<entity-type>")
		return
	}

	entityType := nextPair.Value.(StringCodePairValue).Value
	entity, ok := createEntity(entityType)
	if !ok {
		// swallow unsupported entity
		nextPair, error = reader.readCodePair()
		for error == nil && nextPair.Code != 0 {
			nextPair, error = reader.readCodePair()
		}

		created = false
		return
	}

	created = true
	nextPair, error = reader.readCodePair()
	for error == nil && nextPair.Code != 0 {
		entity.tryApplyCodePair(nextPair)
		nextPair, error = reader.readCodePair()
	}

	afterRead(&entity)
	switch dim := entity.(type) {
	case *dimensionHelper:
		entity, error = createAndPopulateDimension(dim)
	}
	return
}

func writeEntitiesSection(entities []Entity, writer codePairWriter, version AcadVersion) error {
	pairs := make([]CodePair, 0)
	for _, entity := range entities {
		if version >= entity.minVersion() && version <= entity.maxVersion() {
			beforeWrite(&entity)
			for _, pair := range entity.codePairs(version) {
				pairs = append(pairs, pair)
			}
			for _, pair := range trailingCodePairs(&entity, version) {
				pairs = append(pairs, pair)
			}
		}
	}

	err := writeSectionStart(writer, "ENTITIES")
	if err != nil {
		return err
	}
	for _, pair := range pairs {
		err = writer.writeCodePair(pair)
		if err != nil {
			return err
		}
	}
	err = writeSectionEnd(writer)
	if err != nil {
		return err
	}

	return nil
}

func beforeWrite(entity *Entity) {
	switch ent := (*entity).(type) {
	case *ProxyEntity:
		// gather graphics and entity data into strings
		ent.graphicsDataSize = len(ent.GraphicsData)
		ent.graphicsDataString = bytesToStrings(ent.GraphicsData)
		ent.entityDataSize = len(ent.EntityData)
		ent.entityDataString = bytesToStrings(ent.EntityData)
	}
}

func trailingCodePairs(entity *Entity, version AcadVersion) (pairs []CodePair) {
	switch ent := (*entity).(type) {
	case *Attribute:
		pairs = append(pairs, ent.MText.codePairs(version)...)
	case *AttributeDefinition:
		pairs = append(pairs, ent.MText.codePairs(version)...)
	case *Insert:
		for _, att := range ent.Attributes {
			pairs = append(pairs, att.codePairs(version)...)
		}
		pairs = append(pairs, ent.seqend.codePairs(version)...)
	}

	return
}

func afterRead(entity *Entity) {
	switch ent := (*entity).(type) {
	case *Image:
		minLength := len(ent.clippingVerticesX)
		if len(ent.clippingVerticesY) < minLength {
			minLength = len(ent.clippingVerticesY)
		}
		for i := 0; i < minLength; i++ {
			ent.ClippingVertices = append(ent.ClippingVertices, Point{ent.clippingVerticesX[i], ent.clippingVerticesY[i], 0.0})
		}
	case *Leader:
		for i := 0; i < ent.vertexCount; i++ {
			ent.Vertices = append(ent.Vertices, Point{ent.verticesX[i], ent.verticesY[i], ent.verticesZ[i]})
		}
	case *ProxyEntity:
		ent.GraphicsData = stringsToBytes(ent.graphicsDataString)
		ent.EntityData = stringsToBytes(ent.entityDataString)
	}
}

func collectEntities(entityBuffer *entityBufferReader) (result []Entity) {
	for entityBuffer.ItemsRemain() {
		ent := entityBuffer.Peek()
		entityBuffer.Advance()
		switch entity := ent.(type) {
		case *Attribute:
			// ATTRIB should be followed by a single MTEXT
			mtext, err := getNextMText(entityBuffer)
			if err == nil {
				entity.MText = mtext
			}
		case *AttributeDefinition:
			// ATTDEF should be followed by a single MTEXT
			mtext, err := getNextMText(entityBuffer)
			if err == nil {
				entity.MText = mtext
			}
		case *Insert:
			// INSERT should be followed by multiple ATTRIB...
			if entity.HasAttributes {
				for entityBuffer.ItemsRemain() {
					att, err := getNextAttribute(entityBuffer)
					if err == nil {
						entity.Attributes = append(entity.Attributes, att)
					} else {
						break
					}
				}
			}
			// ...and a single SEQEND
			seqend, err := getNextSeqend(entityBuffer)
			if err == nil {
				entity.seqend = seqend
			}
		}

		result = append(result, ent)
	}

	return
}

func getNextAttribute(entityBuffer *entityBufferReader) (att Attribute, error error) {
	if entityBuffer.ItemsRemain() {
		switch next := entityBuffer.Peek().(type) {
		case *Attribute:
			att = *next
			entityBuffer.Advance()
			if entityBuffer.ItemsRemain() {
				switch next := entityBuffer.Peek().(type) {
				case *MText:
					att.MText = *next
					entityBuffer.Advance()
				}
			}
		default:
			error = errors.New("not an attribute")
		}
	} else {
		error = errors.New("no more entities")
	}

	return
}

func getNextMText(entityBuffer *entityBufferReader) (mtext MText, error error) {
	if entityBuffer.ItemsRemain() {
		switch next := entityBuffer.Peek().(type) {
		case *MText:
			mtext = *next
			entityBuffer.Advance()
		default:
			error = errors.New("not an mtext")
		}
	} else {
		error = errors.New("no more entities")
	}

	return
}

func getNextSeqend(entityBuffer *entityBufferReader) (seqend Seqend, error error) {
	if entityBuffer.ItemsRemain() {
		switch next := entityBuffer.Peek().(type) {
		case *Seqend:
			seqend = *next
			entityBuffer.Advance()
		default:
			error = errors.New("not a seqend")
		}
	} else {
		error = errors.New("no more entities")
	}

	return
}

func bytesToStrings(data []byte) []string {
	// for now just return a single large string
	return []string{strings.ToUpper(hex.EncodeToString(data))}
}

func stringsToBytes(vals []string) []byte {
	fullString := strings.Join(vals, "")
	bytes, _ := hex.DecodeString(fullString) // it's ok if this fails
	return bytes
}

//
// temporary dimension helper
//

func (d *dimensionHelper) tryApplyCodePair(codePair CodePair) {
	d.collectedPairs = append(d.collectedPairs, codePair)
}

func (d *dimensionHelper) codePairs(version AcadVersion) (pairs []CodePair) {
	return
}

//
// entity specific methods
//

func (a *Attribute) tryApplyCodePair(codePair CodePair) {
	switch codePair.Code {
	case 100:
		a.lastSubclassMarker = codePair.Value.(StringCodePairValue).Value
	case 1:
		a.Value = codePair.Value.(StringCodePairValue).Value
	case 2:
		if a.lastSubclassMarker == "AcDbXrecord" {
			a.XRecordTag = codePair.Value.(StringCodePairValue).Value
		} else {
			a.AttributeTag = codePair.Value.(StringCodePairValue).Value
		}
	case 7:
		a.TextStyleName = codePair.Value.(StringCodePairValue).Value
	case 10:
		if a.lastSubclassMarker == "AcDbXrecord" {
			a.AlignmentPoint.X = codePair.Value.(DoubleCodePairValue).Value
		} else {
			a.Location.X = codePair.Value.(DoubleCodePairValue).Value
		}
	case 20:
		if a.lastSubclassMarker == "AcDbXrecord" {
			a.AlignmentPoint.Y = codePair.Value.(DoubleCodePairValue).Value
		} else {
			a.Location.Y = codePair.Value.(DoubleCodePairValue).Value
		}
	case 30:
		if a.lastSubclassMarker == "AcDbXrecord" {
			a.AlignmentPoint.Z = codePair.Value.(DoubleCodePairValue).Value
		} else {
			a.Location.Z = codePair.Value.(DoubleCodePairValue).Value
		}
	case 11:
		a.SecondAlignmentPoint.X = codePair.Value.(DoubleCodePairValue).Value
	case 21:
		a.SecondAlignmentPoint.Y = codePair.Value.(DoubleCodePairValue).Value
	case 31:
		a.SecondAlignmentPoint.Z = codePair.Value.(DoubleCodePairValue).Value
	case 39:
		a.Thickness = codePair.Value.(DoubleCodePairValue).Value
	case 40:
		if a.lastSubclassMarker == "AcDbXrecord" {
			a.AnnotationScale = codePair.Value.(DoubleCodePairValue).Value
		} else {
			a.TextHeight = codePair.Value.(DoubleCodePairValue).Value
		}
	case 41:
		a.RelativeXScaleFactor = codePair.Value.(DoubleCodePairValue).Value
	case 50:
		a.Rotation = codePair.Value.(DoubleCodePairValue).Value
	case 51:
		a.ObliqueAngle = codePair.Value.(DoubleCodePairValue).Value
	case 70:
		if a.lastSubclassMarker == "AcDbXrecord" {
			switch a.xrecCode70Count {
			case 0:
				a.MTextFlag = MTextFlag(codePair.Value.(ShortCodePairValue).Value)
			case 1:
				a.IsReallyLocked = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
			case 2:
				a.secondaryAttributeCount = int(codePair.Value.(ShortCodePairValue).Value)
			default:
				// return error?
			}
			a.xrecCode70Count++
		} else {
			a.Flags = int(codePair.Value.(ShortCodePairValue).Value)
		}
	case 71:
		a.TextGenerationFlags = int(codePair.Value.(ShortCodePairValue).Value)
	case 72:
		a.HorizontalTextJustification = HorizontalTextJustification(codePair.Value.(ShortCodePairValue).Value)
	case 73:
		a.FieldLength = codePair.Value.(ShortCodePairValue).Value
	case 74:
		a.VerticalTextJustification = VerticalTextJustification(codePair.Value.(ShortCodePairValue).Value)
	case 210:
		a.Normal.X = codePair.Value.(DoubleCodePairValue).Value
	case 220:
		a.Normal.Y = codePair.Value.(DoubleCodePairValue).Value
	case 230:
		a.Normal.Z = codePair.Value.(DoubleCodePairValue).Value
	case 280:
		if a.lastSubclassMarker == "AcDbXrecord" {
			a.KeepDuplicateRecords = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
		} else if !a.isVersionSet {
			a.Version = Version(codePair.Value.(ShortCodePairValue).Value)
			a.isVersionSet = true
		} else {
			a.IsLockedInBlock = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
		}
	case 340:
		a.secondaryAttributeHandles = append(a.secondaryAttributeHandles, codePair.Value.(StringCodePairValue).Value)
	default:
		tryApplyCodePairForEntity(a, codePair)
	}
}

func (ad *AttributeDefinition) tryApplyCodePair(codePair CodePair) {
	switch codePair.Code {
	case 100:
		ad.lastSubclassMarker = codePair.Value.(StringCodePairValue).Value
	case 1:
		ad.Value = codePair.Value.(StringCodePairValue).Value
	case 2:
		if ad.lastSubclassMarker == "AcDbXrecord" {
			ad.XRecordTag = codePair.Value.(StringCodePairValue).Value
		} else {
			ad.TextTag = codePair.Value.(StringCodePairValue).Value
		}
	case 3:
		ad.Prompt = codePair.Value.(StringCodePairValue).Value
	case 7:
		ad.TextStyleName = codePair.Value.(StringCodePairValue).Value
	case 10:
		if ad.lastSubclassMarker == "AcDbXrecord" {
			ad.AlignmentPoint.X = codePair.Value.(DoubleCodePairValue).Value
		} else {
			ad.Location.X = codePair.Value.(DoubleCodePairValue).Value
		}
	case 20:
		if ad.lastSubclassMarker == "AcDbXrecord" {
			ad.AlignmentPoint.Y = codePair.Value.(DoubleCodePairValue).Value
		} else {
			ad.Location.Y = codePair.Value.(DoubleCodePairValue).Value
		}
	case 30:
		if ad.lastSubclassMarker == "AcDbXrecord" {
			ad.AlignmentPoint.Z = codePair.Value.(DoubleCodePairValue).Value
		} else {
			ad.Location.Z = codePair.Value.(DoubleCodePairValue).Value
		}
	case 11:
		ad.SecondAlignmentPoint.X = codePair.Value.(DoubleCodePairValue).Value
	case 21:
		ad.SecondAlignmentPoint.Y = codePair.Value.(DoubleCodePairValue).Value
	case 31:
		ad.SecondAlignmentPoint.Z = codePair.Value.(DoubleCodePairValue).Value
	case 39:
		ad.Thickness = codePair.Value.(DoubleCodePairValue).Value
	case 40:
		if ad.lastSubclassMarker == "AcDbXrecord" {
			ad.AnnotationScale = codePair.Value.(DoubleCodePairValue).Value
		} else {
			ad.TextHeight = codePair.Value.(DoubleCodePairValue).Value
		}
	case 41:
		ad.RelativeXScaleFactor = codePair.Value.(DoubleCodePairValue).Value
	case 50:
		ad.Rotation = codePair.Value.(DoubleCodePairValue).Value
	case 51:
		ad.ObliqueAngle = codePair.Value.(DoubleCodePairValue).Value
	case 70:
		if ad.lastSubclassMarker == "AcDbXrecord" {
			switch ad.xrecCode70Count {
			case 0:
				ad.MTextFlag = MTextFlag(codePair.Value.(ShortCodePairValue).Value)
			case 1:
				ad.IsReallyLocked = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
			case 2:
				ad.secondaryAttributeHandlesCount = int(codePair.Value.(ShortCodePairValue).Value)
			default:
				// return error?
			}
			ad.xrecCode70Count++
		} else {
			ad.Flags = int(codePair.Value.(ShortCodePairValue).Value)
		}
	case 71:
		ad.TextGenerationFlags = int(codePair.Value.(ShortCodePairValue).Value)
	case 72:
		ad.HorizontalTextJustification = HorizontalTextJustification(codePair.Value.(ShortCodePairValue).Value)
	case 73:
		ad.FieldLength = codePair.Value.(ShortCodePairValue).Value
	case 74:
		ad.VerticalTextJustification = VerticalTextJustification(codePair.Value.(ShortCodePairValue).Value)
	case 210:
		ad.Normal.X = codePair.Value.(DoubleCodePairValue).Value
	case 220:
		ad.Normal.Y = codePair.Value.(DoubleCodePairValue).Value
	case 230:
		ad.Normal.Z = codePair.Value.(DoubleCodePairValue).Value
	case 280:
		if ad.lastSubclassMarker == "AcDbXrecord" {
			ad.KeepDuplicateRecords = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
		} else if !ad.isVersionSet {
			ad.Version = Version(codePair.Value.(ShortCodePairValue).Value)
			ad.isVersionSet = true
		} else {
			ad.IsLockedInBlock = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
		}
	case 340:
		ad.SecondaryAttributeHandles = append(ad.SecondaryAttributeHandles, codePair.Value.(StringCodePairValue).Value)
	default:
		tryApplyCodePairForEntity(ad, codePair)
	}
}

func (pl *LWPolyline) tryApplyCodePair(codePair CodePair) {
	switch codePair.Code {
	// entity codes
	case 39:
		pl.Thickness = codePair.Value.(DoubleCodePairValue).Value
	case 43:
		pl.ConstantWidth = codePair.Value.(DoubleCodePairValue).Value
	case 70:
		pl.flags = int(codePair.Value.(ShortCodePairValue).Value)
	case 90:
		pl.vertexCount = codePair.Value.(IntCodePairValue).Value
	case 210:
		pl.ExtrusionDirection.X = codePair.Value.(DoubleCodePairValue).Value
	case 220:
		pl.ExtrusionDirection.Y = codePair.Value.(DoubleCodePairValue).Value
	case 230:
		pl.ExtrusionDirection.Z = codePair.Value.(DoubleCodePairValue).Value
	// vertex codes
	case 10:
		// start a new vertex
		v := *NewVertex()
		v.X = codePair.Value.(DoubleCodePairValue).Value
		pl.Vertices = append(pl.Vertices, v)
	case 20:
		// update the last vertex
		if len(pl.Vertices) > 0 {
			pl.Vertices[len(pl.Vertices)-1].Y = codePair.Value.(DoubleCodePairValue).Value
		}
	case 40:
		// update the last vertex
		if len(pl.Vertices) > 0 {
			pl.Vertices[len(pl.Vertices)-1].StartingWidth = codePair.Value.(DoubleCodePairValue).Value
		}
	case 41:
		// update the last vertex
		if len(pl.Vertices) > 0 {
			pl.Vertices[len(pl.Vertices)-1].EndingWidth = codePair.Value.(DoubleCodePairValue).Value
		}
	case 42:
		// update the last vertex
		if len(pl.Vertices) > 0 {
			pl.Vertices[len(pl.Vertices)-1].Bulge = codePair.Value.(DoubleCodePairValue).Value
		}
	case 91:
		// update the last vertex
		if len(pl.Vertices) > 0 {
			pl.Vertices[len(pl.Vertices)-1].ID = codePair.Value.(IntCodePairValue).Value
		}
	default:
		appliedCodePair := false
		if !appliedCodePair {
			appliedCodePair = tryApplyCodePairForEntity(pl, codePair)
		}
	}
}

func (mt *MText) tryApplyCodePair(codePair CodePair) {
	switch codePair.Code {
	case 10:
		mt.InsertionPoint.X = codePair.Value.(DoubleCodePairValue).Value
	case 20:
		mt.InsertionPoint.Y = codePair.Value.(DoubleCodePairValue).Value
	case 30:
		mt.InsertionPoint.Z = codePair.Value.(DoubleCodePairValue).Value
	case 40:
		mt.InitialTextHeight = codePair.Value.(DoubleCodePairValue).Value
	case 41:
		mt.ReferenceRectangleWidth = codePair.Value.(DoubleCodePairValue).Value
	case 71:
		mt.AttachmentPoint = AttachmentPoint(codePair.Value.(ShortCodePairValue).Value)
	case 72:
		mt.DrawingDirection = DrawingDirection(codePair.Value.(ShortCodePairValue).Value)
	case 3:
		mt.ExtendedText = append(mt.ExtendedText, codePair.Value.(StringCodePairValue).Value)
	case 1:
		mt.Text = codePair.Value.(StringCodePairValue).Value
	case 7:
		mt.TextStyleName = codePair.Value.(StringCodePairValue).Value
	case 210:
		mt.ExtrusionDirection.X = codePair.Value.(DoubleCodePairValue).Value
	case 220:
		mt.ExtrusionDirection.Y = codePair.Value.(DoubleCodePairValue).Value
	case 230:
		mt.ExtrusionDirection.Z = codePair.Value.(DoubleCodePairValue).Value
	case 11:
		mt.XAxisDirection.X = codePair.Value.(DoubleCodePairValue).Value
	case 21:
		mt.XAxisDirection.Y = codePair.Value.(DoubleCodePairValue).Value
	case 31:
		mt.XAxisDirection.Z = codePair.Value.(DoubleCodePairValue).Value
	case 42:
		mt.HorizontalWidth = codePair.Value.(DoubleCodePairValue).Value
	case 43:
		mt.VerticalHeight = codePair.Value.(DoubleCodePairValue).Value
	case 50:
		if mt.readingColumnData {
			if mt.readColumnCount {
				mt.ColumnHeights = append(mt.ColumnHeights, codePair.Value.(DoubleCodePairValue).Value)
			} else {
				mt.columnCount = int16(codePair.Value.(DoubleCodePairValue).Value)
				mt.readColumnCount = true
			}
		} else {
			mt.RotationAngle = codePair.Value.(DoubleCodePairValue).Value
		}
	case 73:
		mt.LineSpacingStyle = MTextLineSpacingStyle(codePair.Value.(ShortCodePairValue).Value)
	case 44:
		mt.LineSpacingFactor = codePair.Value.(DoubleCodePairValue).Value
	case 90:
		mt.BackgroundFillSetting = BackgroundFillSetting(codePair.Value.(IntCodePairValue).Value)
	case 420:
		mt.BackgroundColorRGB = codePair.Value.(IntCodePairValue).Value
	case 430:
		mt.BackgroundColorName = codePair.Value.(StringCodePairValue).Value
	case 45:
		mt.FillBoxScale = codePair.Value.(DoubleCodePairValue).Value
	case 63:
		mt.BackgroundFillColor = Color(codePair.Value.(ShortCodePairValue).Value)
	case 441:
		mt.BackgroundFillColorTransparency = codePair.Value.(IntCodePairValue).Value
	case 75:
		mt.ColumnType = codePair.Value.(ShortCodePairValue).Value
		mt.readingColumnData = true
	case 76:
		mt.columnCount = codePair.Value.(ShortCodePairValue).Value
	case 78:
		mt.IsColumnFlowReversed = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
	case 79:
		mt.IsColumnAutoHeight = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
	case 48:
		mt.ColumnWidth = codePair.Value.(DoubleCodePairValue).Value
	case 49:
		mt.ColumnGutter = codePair.Value.(DoubleCodePairValue).Value
	default:
		tryApplyCodePairForEntity(mt, codePair)
	}
}

func (entity *ProxyEntity) tryApplyCodePair(codePair CodePair) {
	switch codePair.Code {
	case 90:
		entity.ProxyEntityClassId = codePair.Value.(IntCodePairValue).Value
	case 91:
		entity.ApplicationEntityClassId = codePair.Value.(IntCodePairValue).Value
	case 92:
		entity.graphicsDataSize = codePair.Value.(IntCodePairValue).Value
	case 310:
		if entity.readingGraphicsData {
			entity.graphicsDataString = append(entity.graphicsDataString, codePair.Value.(StringCodePairValue).Value)
		} else {
			entity.entityDataString = append(entity.entityDataString, codePair.Value.(StringCodePairValue).Value)
		}
	case 93:
		entity.entityDataSize = codePair.Value.(IntCodePairValue).Value
		entity.readingGraphicsData = false
	case 330:
		entity.ObjectID1 = append(entity.ObjectID1, codePair.Value.(StringCodePairValue).Value)
	case 340:
		entity.ObjectID2 = append(entity.ObjectID2, codePair.Value.(StringCodePairValue).Value)
	case 350:
		entity.ObjectID3 = append(entity.ObjectID3, codePair.Value.(StringCodePairValue).Value)
	case 360:
		entity.ObjectID4 = append(entity.ObjectID4, codePair.Value.(StringCodePairValue).Value)
	case 94:
		entity.Terminator = codePair.Value.(IntCodePairValue).Value
	case 95:
		entity.objectDrawingFormat = uint(codePair.Value.(IntCodePairValue).Value)
	case 70:
		entity.OriginalDataFormatIsDxf = boolFromShort(codePair.Value.(ShortCodePairValue).Value)
	default:
		tryApplyCodePairForEntity(entity, codePair)
	}
}

func (s *Seqend) tryApplyCodePair(codePair CodePair) {
	tryApplyCodePairForEntity(s, codePair)
}
