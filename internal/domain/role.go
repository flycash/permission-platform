package domain

// Role 角色
type Role struct {
	ID          int64
	BizID       int64
	Type        string
	Name        string
	Description string
	Metadata    string
	Ctime       int64
	Utime       int64
}
