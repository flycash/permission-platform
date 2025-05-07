package dao

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
