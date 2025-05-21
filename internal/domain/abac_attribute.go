package domain

import (
	"gitee.com/flycash/permission-platform/internal/errs"
	"github.com/ecodeclub/ekit/slice"
)

// 业务的定义
type BizAttrDefinition struct {
	BizID               int64
	SubjectAttrDefs     AttrDefs // 主体的属性定义
	ResourceAttrDefs    AttrDefs // 资源的属性定义
	EnvironmentAttrDefs AttrDefs // 环境的属性定义
	AllDefs             map[int64]AttributeDefinition
}

func (biz BizAttrDefinition) GetByDefID(id int64) (AttributeDefinition, bool) {
	def, ok := biz.AllDefs[id]
	return def, ok
}

type (
	AttributeDataType string
	AttributeType     string
)

const (
	SubjectType     AttributeType = "subject"
	EnvironmentType AttributeType = "environment"
	ResourceType    AttributeType = "resource"
)

func (t AttributeType) String() string {
	return string(t)
}

type AttrDefs []AttributeDefinition

func (a AttrDefs) Map() map[int64]AttributeDefinition {
	res := make(map[int64]AttributeDefinition, len(a))
	for idx := range a {
		val := a[idx]
		res[val.ID] = val
	}
	return res
}

func (a AttrDefs) GetDefinition(id int64) (AttributeDefinition, bool) {
	for idx := range a {
		if a[idx].ID == id {
			return a[idx], true
		}
	}
	return AttributeDefinition{}, false
}

func (a AttrDefs) GetDefinitionWithName(name string) (AttributeDefinition, bool) {
	for idx := range a {
		if a[idx].Name == name {
			return a[idx], true
		}
	}
	return AttributeDefinition{}, false
}

func (s *ABACObject) ValuesMap() map[int64]string {
	return slice.ToMapV(s.AttributeValues, func(element AttributeValue) (int64, string) {
		return element.ID, element.Value
	})
}

type AttributeValue struct {
	ID         int64
	Definition AttributeDefinition // 对应的属性定义
	Value      string              // 对应的属性值
	Ctime      int64
	Utime      int64
}

// 资源对象
type ABACObject struct {
	BizID           int64
	ID              int64
	AttributeValues []AttributeValue
}

func (s *ABACObject) MergeRealTimeAttrs(attrs AttrDefs, values map[string]string) {
	for idx := range s.AttributeValues {
		val := s.AttributeValues[idx]
		def, _ := attrs.GetDefinition(val.Definition.ID)
		s.AttributeValues[idx].Definition = def
		s.AttributeValues[idx].Value = values[def.Name]
	}
}

func (s *ABACObject) AttributeVal(attributeID int64) (AttributeValue, error) {
	for idx := range s.AttributeValues {
		if s.AttributeValues[idx].Definition.ID == attributeID {
			return s.AttributeValues[idx], nil
		}
	}
	return AttributeValue{}, errs.ErrAttributeNotFound
}

func (s *ABACObject) SetAttributeVal(val string, definition AttributeDefinition) {
	for idx := range s.AttributeValues {
		if s.AttributeValues[idx].Definition.ID == definition.ID {
			s.AttributeValues[idx].Value = val
			return
		}
	}
	s.AttributeValues = append(s.AttributeValues, AttributeValue{
		Definition: definition,
		Value:      val,
	})
}

type Attributes struct {
	Subject     map[string]string // 属性名 name => 属性值 value
	Resource    map[string]string
	Environment map[string]string
}
