package influxqu

type UnSupportedType struct{}

func (e *UnSupportedType) Error() string {
	return "unsupported type"
}

type DuplicatedMeasurement struct{}

func (e *DuplicatedMeasurement) Error() string {
	return "duplicated measurement"
}

type DuplicatedTimestamp struct{}

func (e *DuplicatedTimestamp) Error() string {
	return "duplicated timestamp"
}

type DuplicatedTag struct {
	tag string
}

func (e *DuplicatedTag) Error() string {
	return "duplicated tag " + e.tag
}

type DuplicatedField struct {
	field string
}

func (e *DuplicatedField) Error() string {
	return "duplicated field " + e.field
}

type NoTagName struct{}

func (e *NoTagName) Error() string {
	return "no tag name"
}

type NoFieldName struct{}

func (e *NoFieldName) Error() string {
	return "no field name"
}

type DuplicatedKey struct{}

func (e *DuplicatedKey) Error() string {
	return "duplicated key"
}

type UnSupportedTag struct{}

func (e *UnSupportedTag) Error() string {
	return "unsupported tag"
}
