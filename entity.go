package dxf

import (
	"encoding/hex"
	"errors"
	"strings"
)

func readEntity(nextPair CodePair, reader codePairReader) (Entity, CodePair, bool, error) {
	var entity Entity
	if nextPair.Code != 0 {
		return entity, nextPair, false, errors.New("exepcted 0/<entity-type>")
	}

	var err error
	entityType := nextPair.Value.(StringCodePairValue).Value
	entity, ok := createEntity(entityType)
	if !ok {
		// swallow unsupported entity
		nextPair, err = reader.readCodePair()
		for err == nil && nextPair.Code != 0 {
			nextPair, err = reader.readCodePair()
		}

		return entity, nextPair, false, nil
	}

	nextPair, err = reader.readCodePair()
	for err == nil && nextPair.Code != 0 {
		entity.tryApplyCodePair(nextPair)
		nextPair, err = reader.readCodePair()
	}

	afterRead(&entity)
	return entity, nextPair, true, err
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

func afterRead(entity *Entity) {
	switch ent := (*entity).(type) {
	case *ProxyEntity:
		ent.GraphicsData = stringsToBytes(ent.graphicsDataString)
		ent.EntityData = stringsToBytes(ent.entityDataString)
	}
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
// entity specific methods
//

func (entity *ProxyEntity) tryApplyCodePair(codePair CodePair) {
	switch codePair.Code {
	// entity specific values
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
		tryApplyBaseCodePair(entity, codePair)
	}
}
