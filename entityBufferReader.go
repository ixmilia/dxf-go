package dxf

type entityBufferReader struct {
	entities []Entity
	position int
}

func (reader *entityBufferReader) ItemsRemain() bool {
	return reader.position < len(reader.entities)
}

func (reader *entityBufferReader) Peek() Entity {
	return reader.entities[reader.position]
}

func (reader *entityBufferReader) Advance() {
	reader.position++
}
