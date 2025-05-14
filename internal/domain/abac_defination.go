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

// 负责转化和比较，用户需要提供对应类型的转化方式，和比较方法
// case1: actualVal = 20 wantval = []int
// case2: actualval = [1,3,5] wantval = [1,2,3,4,5] 不支持 anymatch
