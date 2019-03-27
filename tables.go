package dxf

import "errors"

func readTables(drawing *Drawing, np CodePair, reader codePairReader) (nextPair CodePair, error error) {
	nextPair = np
	for error == nil && !nextPair.isEndSection() {
		if !nextPair.isStartTable() {
			error = errors.New("expected 0/TABLE")
			return
		}
		nextPair, error = reader.readCodePair()
		if error != nil {
			return
		}
		// swallow until 2/<table-type>
		for error == nil && nextPair.Code != 2 {
			nextPair, error = reader.readCodePair()
		}
		if error != nil {
			return
		}
		tableType := nextPair.Value.(StringCodePairValue).Value
		nextPair, error = reader.readCodePair()
		if error != nil {
			return
		}
		nextPair, error = readSpecificTable(drawing, nextPair, reader, tableType)
		// swallow until 0/ENDTAB
		for error == nil && !nextPair.isEndTable() {
			nextPair, error = reader.readCodePair()
		}
		if error != nil {
			return
		}

		// swallow the actual 0/ENDTAB
		nextPair, error = reader.readCodePair()
	}
	return
}

func writeTablesSection(drawing *Drawing, writer codePairWriter, version AcadVersion) (err error) {
	pairs := getTablePairs(drawing, version)

	err = writeSectionStart(writer, "TABLES")
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
	return
}
