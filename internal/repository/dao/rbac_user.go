package dao

// EffectType 权限效果枚举
type EffectType string

const (
	EffectTypeAllow EffectType = "allow" // 允许
	EffectTypeDeny  EffectType = "deny"  // 拒绝
)

// UserRole 用户角色关联关系表
type UserRole struct {
	ID        int64    `gorm:"primaryKey;autoIncrement;comment:用户角色关联关系主键'"`
	BizID     int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:1;index:idx_biz_user,priority:1;index:idx_biz_role,priority:1;index:idx_biz_user_role_validity,priority:1;comment:'业务ID'"`
	UserID    int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:2;index:idx_biz_user,priority:2;index:idx_biz_user_role_validity,priority:2;comment:'用户ID'"`
	RoleID    int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_role,priority:3;index:idx_biz_role,priority:2;comment:'角色ID'"`
	RoleName  string   `gorm:"type:VARCHAR(100);NOT NULL;comment:'角色名称（冗余字段，加速查询）'"`
	RoleType  RoleType `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;index:idx_biz_user_role_validity,priority:3;comment:'角色类型（冗余字段，加速查询）'"`
	StartTime int64    `gorm:"NULL;index:idx_biz_user_role_validity,priority:4;comment:'临时角色生效时间（冗余字段，加速查询）'"`
	EndTime   int64    `gorm:"NULL;index:idx_biz_user_role_validity,priority:5;comment:'临时角色失效时间（冗余字段，加速查询）'"`
	Ctime     int64
	Utime     int64
}

func (UserRole) TableName() string {
	return "user_roles"
}

// UserPermission 用户个人权限关联表
type UserPermission struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;comment:'用户权限关联关系ID'"`
	BizID          int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:1;index:idx_biz_user,priority:1;index:idx_biz_permission,priority:1;index:idx_biz_effect,priority:1;index:idx_biz_resource_type,priority:1;index:idx_biz_action,priority:1;index:idx_time_range,priority:1;index:idx_current_valid,priority:1;index:idx_biz_resource_key_action,priority:1;comment:'业务ID'"`
	UserID         int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:2;index:idx_biz_user,priority:2;comment:'用户ID'"`
	PermissionID   int64      `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_user_permission,priority:3;index:idx_biz_permission,priority:2;comment:'权限ID'"`
	PermissionName string     `gorm:"type:VARCHAR(100);NOT NULL;comment:'权限名称（冗余字段，加速查询与展示）'"`
	ResourceType   string     `gorm:"type:VARCHAR(100);NOT NULL;index:idx_biz_resource_type,priority:2;index:idx_biz_resource_key_action,priority:2;comment:'资源类型（冗余字段，加速查询）'"`
	ResourceKey    string     `gorm:"type:VARCHAR(255);NOT NULL;index:idx_biz_resource_key_action,priority:3;comment:'资源标识符（冗余字段，加速查询）'"`
	ResourceName   string     `gorm:"type:VARCHAR(255);NOT NULL;comment:'资源名称（冗余字段，加速查询与展示）'"`
	Action         ActionType `gorm:"type:ENUM('create', 'read', 'update', 'delete', 'execute', 'export', 'import');NOT NULL;index:idx_biz_action,priority:2;index:idx_biz_resource_key_action,priority:4;comment:'操作类型（冗余字段，加速查询）'"`
	StartTime      int64      `gorm:"NULL;index:idx_time_range,priority:2;index:idx_current_valid,priority:3;comment:'权限生效时间，如果设置了表示临时权限'"`
	EndTime        int64      `gorm:"NULL;index:idx_time_range,priority:3;index:idx_current_valid,priority:4;comment:'权限失效时间，如果设置了表示临时权限'"`
	Effect         EffectType `gorm:"type:ENUM('allow', 'deny');NOT NULL;DEFAULT:'allow';index:idx_biz_effect,priority:2;index:idx_current_valid,priority:2;comment:'用于额外授予权限，或者取消权限，理论上不应该出现同时allow和deny，出现了就是deny优先于allow'"`
	Ctime          int64
	Utime          int64
}

func (UserPermission) TableName() string {
	return "user_permissions"
}
