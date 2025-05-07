package dao

// RoleInclusion 角色包含关系表
type RoleInclusion struct {
	ID                int64    `gorm:"primaryKey;autoIncrement;comment:角色包含关系ID'"`
	BizID             int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:1;index:idx_biz_including_role,priority:1;index:idx_biz_included_role,priority:1;comment:'业务ID'"`
	IncludingRoleID   int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:2;index:idx_biz_including_role,priority:2;comment:'包含者角色ID（拥有其他角色权限）'"`
	IncludingRoleType RoleType `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;comment:'包含者角色类型（冗余字段，加速查询）'"`
	IncludingRoleName string   `gorm:"type:VARCHAR(100);NOT NULL;comment:'包含者角色名称（冗余字段，加速查询）'"`
	IncludedRoleID    int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:3;index:idx_biz_included_role,priority:2;comment:'被包含角色ID（权限被包含）'"`
	IncludedRoleType  RoleType `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;comment:'被包含角色类型（冗余字段，加速查询）'"`
	IncludedRoleName  string   `gorm:"type:VARCHAR(100);NOT NULL;comment:'被包含角色名称（冗余字段，加速查询）'"`
	Ctime             int64
	Utime             int64
}

func (RoleInclusion) TableName() string {
	return "role_inclusions"
}
