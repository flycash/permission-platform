package domain

import "encoding/json"

// PermissionMetadata 权限元数据
type PermissionMetadata struct{}

func (m PermissionMetadata) String() string {
	marshal, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(marshal)
}

type Permission struct {
	ID          int64
	BizID       int64
	Name        string
	Description string
	Resource    Resource
	Action      string
	Metadata    PermissionMetadata
	Ctime       int64
	Utime       int64
}
