package dxf

import "errors"

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

	return entity, nextPair, true, err
}
