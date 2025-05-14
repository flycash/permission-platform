package domain

import (
	"gitee.com/flycash/permission-platform/internal/errs"
)

// 业务的定义
type BizDefinition struct {
	BizID            int64
	SubjectAttrs     Attrs // 主体的属性定义
	ResourceAttrs    Attrs // 资源的属性定义
	EnvironmentAttrs Attrs // 环境的属性定义
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

type Attrs []AttributeDefinition

func (a Attrs) Map() map[int64]AttributeDefinition {
	res := make(map[int64]AttributeDefinition, len(a))
	for idx := range a {
		val := a[idx]
		res[val.ID] = val
	}
	return res
}

func (a Attrs) GetDefinition(id int64) (AttributeDefinition, bool) {
	for idx := range a {
		if a[idx].ID == id {
			return a[idx], true
		}
	}
	return AttributeDefinition{}, false
}

func (a Attrs) GetDefinitionWithName(name string) (AttributeDefinition, bool) {
	for idx := range a {
		if a[idx].Name == name {
			return a[idx], true
		}
	}
	return AttributeDefinition{}, false
}

// 具体属性的定义
// 主体对象
type SubjectObject struct {
	BizID           int64
	ID              int64
	AttributeValues []SubjectAttributeValue // 主体对应的属性
}

func (s *SubjectObject) AttributeVal(attributeID int64) (SubjectAttributeValue, error) {
	for idx := range s.AttributeValues {
		if s.AttributeValues[idx].Definition.ID == attributeID {
			return s.AttributeValues[idx], nil
		}
	}
	return SubjectAttributeValue{}, errs.ErrAttributeNotFound
}

func (s *SubjectObject) SetAttributeVal(val string, definition AttributeDefinition) {
	for idx := range s.AttributeValues {
		if s.AttributeValues[idx].Definition.ID == definition.ID {
			s.AttributeValues[idx].Value = val
			return
		}
	}
	s.AttributeValues = append(s.AttributeValues, SubjectAttributeValue{
		Definition: definition,
		Value:      val,
	})
}

type SubjectAttributeValue struct {
	ID         int64
	Definition AttributeDefinition // 对应的属性定义
	Value      string              // 对应的属性值
	Ctime      int64
	Utime      int64
}

// 资源对象
type ResourceObject struct {
	BizID           int64
	ID              int64
	AttributeValues []ResourceAttributeValue
}

type ResourceAttributeValue struct {
	ID         int64
	Definition AttributeDefinition
	Value      string
	Ctime      int64
	Utime      int64
}

func (s *ResourceObject) AttributeVal(attributeID int64) (ResourceAttributeValue, error) {
	for idx := range s.AttributeValues {
		if s.AttributeValues[idx].Definition.ID == attributeID {
			return s.AttributeValues[idx], nil
		}
	}
	return ResourceAttributeValue{}, errs.ErrAttributeNotFound
}

func (s *ResourceObject) SetAttributeVal(val string, definition AttributeDefinition) {
	for idx := range s.AttributeValues {
		if s.AttributeValues[idx].Definition.ID == definition.ID {
			s.AttributeValues[idx].Value = val
			return
		}
	}
	s.AttributeValues = append(s.AttributeValues, ResourceAttributeValue{
		Definition: definition,
		Value:      val,
	})
}

// 环境对象
type EnvironmentObject struct {
	BizID           int64
	AttributeValues []EnvironmentAttributeValue
}

type EnvironmentAttributeValue struct {
	ID         int64
	Definition AttributeDefinition
	Value      string
	Ctime      int64
	Utime      int64
}

func (s *EnvironmentObject) AttributeVal(attributeID int64) (EnvironmentAttributeValue, error) {
	for idx := range s.AttributeValues {
		if s.AttributeValues[idx].Definition.ID == attributeID {
			return s.AttributeValues[idx], nil
		}
	}
	return EnvironmentAttributeValue{}, errs.ErrAttributeNotFound
}

func (s *EnvironmentObject) SetAttributeVal(val string, definition AttributeDefinition) {
	for idx := range s.AttributeValues {
		if s.AttributeValues[idx].Definition.ID == definition.ID {
			s.AttributeValues[idx].Value = val
			return
		}
	}
	s.AttributeValues = append(s.AttributeValues, EnvironmentAttributeValue{
		Definition: definition,
		Value:      val,
	})
}
