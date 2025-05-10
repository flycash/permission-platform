package domain

type ActionType string

const (
	ActionTypeCreate  ActionType = "create"
	ActionTypeRead    ActionType = "read"
	ActionTypeWrite   ActionType = "write"
	ActionTypeDelete  ActionType = "delete"
	ActionTypeExecute ActionType = "execute"
	ActionTypeExport  ActionType = "export"
	ActionTypeImport  ActionType = "import"
)

// PermissionMetadata 权限元数据
type PermissionMetadata struct {
	FilterScope  string `json:"filterScope,omitempty"`  // 过滤范围: all, self, self_and_children
	FilterParams []any  `json:"filterParams,omitempty"` // 过滤参数（通常是bizID列表）
}

type Permission struct {
	ID           int64      `json:"id"`
	BizID        int64      `json:"bizId"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	ResourceID   int64      `json:"resourceId"`
	ResourceType string     `json:"resourceType"`
	ResourceKey  string     `json:"resourceKey"`
	Action       ActionType `json:"action"`
	Ctime        int64
	Utime        int64
}
