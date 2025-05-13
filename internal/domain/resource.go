package domain

import "encoding/json"

// ResourceMetadata 资源元数据
type ResourceMetadata struct{}

func (m ResourceMetadata) String() string {
	marshal, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(marshal)
}

type Resource struct {
	ID          int64
	BizID       int64
	Type        string
	Key         string
	Name        string
	Description string
	Metadata    ResourceMetadata
	Ctime       int64
	Utime       int64
}
