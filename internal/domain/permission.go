package domain

type Permission struct {
	ID          int64
	BizID       int64
	Name        string
	Description string
	Resource    Resource
	Action      string
	Metadata    string
	Ctime       int64
	Utime       int64
}
