package domain

// UserRole 用户角色关联
type UserRole struct {
	ID        int64
	BizID     int64
	UserID    int64
	Role      Role
	StartTime int64 // 授予角色生效时间
	EndTime   int64 // 授予角色失效时间
	Ctime     int64
	Utime     int64
}
