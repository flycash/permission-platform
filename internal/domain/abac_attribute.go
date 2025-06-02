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

func (a AttrDefs) GetByID(id int64) (AttributeDefinition, bool) {
	for idx := range a {
		if a[idx].ID == id {
			return a[idx], true
		}
	}
	return AttributeDefinition{}, false
}

func (a AttrDefs) GetByName(name string) (AttributeDefinition, bool) {
	for idx := range a {
		if a[idx].Name == name {
			return a[idx], true
		}
	}
	return AttributeDefinition{}, false
}

type AttributeValue struct {
	ID         int64
	Definition AttributeDefinition // 对应的属性定义
	Value      string              // 对应的属性值
	Ctime      int64
	Utime      int64
}

// ABACObject ABAC 的对象
type ABACObject struct {
	BizID           int64 `json:"bizId"`
	ID              int64 `json:"id"`
	AttributeValues []AttributeValue
}

func (s *ABACObject) ValuesMap() map[int64]AttributeValue {
	return slice.ToMapV(s.AttributeValues, func(element AttributeValue) (int64, AttributeValue) {
		return element.Definition.ID, element
	})
}

func (s *ABACObject) MergeRealTimeAttrs(attrs AttrDefs, values map[string]string) {
	// 这里是用实时计算的覆盖了存储的
	for key, val := range values {
		def, ok := attrs.GetByName(key)
		if ok {
			s.SetAttributeVal(val, def)
		}
		// 如果不 OK，就是业务方传了一个属性，但是这个属性都不是我们内部的属性
	}
}

func (s *ABACObject) FillDefinitions(attrs AttrDefs) {
	// 这是预存的属性
	for idx := range s.AttributeValues {
		val := s.AttributeValues[idx]
		def, ok := attrs.GetByID(val.Definition.ID)
		if ok {
			s.AttributeValues[idx].Definition = def
		}
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
	Subject     SubAttrs // 属性名 name => 属性值 value
	Resource    SubAttrs
	Environment SubAttrs
}

type SubAttrs map[string]string

func (s SubAttrs) SetKv(k, v string) SubAttrs {
	if s == nil {
		s = map[string]string{
			k: v,
		}
	} else {
		s[k] = v
	}
	return s
}
