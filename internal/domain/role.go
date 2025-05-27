package domain

// Role 角色
type Role struct {
	ID          int64  `json:"id,omitzero"`
	BizID       int64  `json:"bizId,omitzero"`
	Type        string `json:"type,omitzero"`
	Name        string `json:"name,omitzero"`
	Description string `json:"description,omitzero"`
	Metadata    string `json:"metadata,omitzero"`
	Ctime       int64  `json:"ctime,omitzero"`
	Utime       int64  `json:"utime,omitzero"`
}
