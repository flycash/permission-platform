package domain

type Permission struct {
	ID          int64    `json:"id,omitzero"`
	BizID       int64    `json:"bizId,omitzero"`
	Name        string   `json:"name,omitzero"`
	Description string   `json:"description,omitzero"`
	Resource    Resource `json:"resource,omitzero"`
	Action      string   `json:"action,omitzero"`
	Metadata    string   `json:"metadata,omitzero"`
	Ctime       int64    `json:"ctime,omitzero"`
	Utime       int64    `json:"utime,omitzero"`
}
