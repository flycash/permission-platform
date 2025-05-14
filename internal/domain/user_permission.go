package domain

type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

func (e Effect) String() string {
	return string(e)
}

func (e Effect) IsAllow() bool {
	return e.String() == "allow"
}

func (e Effect) IsDeny() bool {
	return e.String() == "deny"
}

// UserPermission 用户权限关联
type UserPermission struct {
	ID         int64
	BizID      int64
	UserID     int64
	Permission Permission
	StartTime  int64 // 权限生效时间
	EndTime    int64 // 权限失效时间
	Effect     Effect
	Ctime      int64
	Utime      int64
}
