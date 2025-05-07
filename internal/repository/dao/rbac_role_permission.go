package dao

// RolePermission 角色权限关联表
type RolePermission struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;comment:'角色权限关联关系ID'"`
	BizID        int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:1;index:idx_biz_role,priority:1;index:idx_biz_permission,priority:1;index:idx_biz_role_type,priority:1;index:idx_biz_resource_type,priority:1;index:idx_biz_action,priority:1;index:idx_biz_resource_key_action,priority:1;comment:'业务ID'"`
	RoleID       int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:2;index:idx_biz_role,priority:2;comment:'角色ID'"`
	PermissionID int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_role_permission,priority:3;index:idx_biz_permission,priority:2;comment:'权限ID'"`
	RoleName     string     `gorm:"type:VARCHAR(100);NOT NULL;comment:'角色名称（冗余字段，加速查询）'"`
	RoleType     RoleType   `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;index:idx_biz_role_type,priority:2;comment:'角色类型（冗余字段，加速查询）'"`
	ResourceType string     `gorm:"type:VARCHAR(100);NOT NULL;index:idx_biz_resource_type,priority:2;index:idx_biz_resource_key_action,priority:2;comment:'资源类型（冗余字段，加速查询）'"`
	ResourceKey  string     `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_key_action,priority:3;comment:'资源标识符（冗余字段，加速查询）'"`
	Action       ActionType `gorm:"type:ENUM('create', 'read', 'update', 'delete', 'execute', 'export', 'import');NOT NULL;index:idx_biz_action,priority:2;index:idx_biz_resource_key_action,priority:4;comment:'操作类型（冗余字段，加速查询）'"`
	Ctime        int64
	Utime        int64
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
