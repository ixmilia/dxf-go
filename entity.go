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

	entity.afterRead()
	return entity, nextPair, true, err
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

// ProxyEntity
func (entity *ProxyEntity) beforeWrite() {
	// gather graphics and entity data into strings
	entity.graphicsDataSize = len(entity.GraphicsData)
	entity.graphicsDataString = bytesToStrings(entity.GraphicsData)
	entity.entityDataSize = len(entity.EntityData)
	entity.entityDataString = bytesToStrings(entity.EntityData)
}

func (entity *ProxyEntity) afterRead() {
	// collect graphics and entity data into byte arrays
	entity.GraphicsData = stringsToBytes(entity.graphicsDataString)
	entity.EntityData = stringsToBytes(entity.entityDataString)
}

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
