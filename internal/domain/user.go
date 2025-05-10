package domain

type User struct {
	ID    int64 `json:"id"`
	BizID int64 `json:"bizId"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID        int64    `json:"id"`
	BizID     int64    `json:"bizId"`
	UserID    int64    `json:"userId"`
	RoleID    int64    `json:"roleId"`
	RoleName  string   `json:"roleName"`
	RoleType  RoleType `json:"roleType"`
	StartTime int64    `json:"startTime"`
	EndTime   int64    `json:"endTime"`
	Ctime     int64
	Utime     int64
}
