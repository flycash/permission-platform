package domain

type AttributeDefinition struct {
	ID             int64
	Name           string
	Description    string
	DataType       DataType
	EntityType     EntityType
	ValidationRule string
	Ctime          int64
	Utime          int64
}

type DataType string

const (
	DataTypeString   DataType = "string"
	DataTypeNumber            = "number"
	DataTypeBoolean           = "boolean"
	DataTypeFloat             = "float"
	DataTypeDatetime          = "datetime"
	DataTypeArray         = "array"
)

func (d DataType) String() string {
	return string(d)
}

type EntityType string

const (
	EntityTypeSubject     EntityType = "subject"
	EntityTypeResource    EntityType = "resource"
	EntityTypeEnvironment EntityType = "environment"
)

func (t EntityType) String() string {
	return string(t)
}
