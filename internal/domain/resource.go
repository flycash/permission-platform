package domain

// ResourceMetadata 资源元数据
type ResourceMetadata struct {
	Owner    int64    `json:"owner,omitempty"`    // 资源所有者
	ParentID int64    `json:"parentId,omitempty"` // 父资源ID（用于树形结构）
	Tags     []string `json:"tags,omitempty"`     // 资源标签
}

// 系统资源键
const (
	RoleTalbeSystemResourceKey       = "permission_system:role"
	ResourceTalbeSystemResourceKey   = "permission_system:resource"
	PermissionTalbeSystemResourceKey = "permission_system:permission"
)

type Resource struct {
	ID          int64  `json:"id"`
	BizID       int64  `json:"bizId"`
	Type        string `json:"type"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Ctime       int64
	Utime       int64
}
