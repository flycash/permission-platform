package gorm

import "context"

// 进行权限校验的model要实现这个接口
type AuthRequired interface {
	ResourceKey(ctx context.Context) string
	ResourceType(ctx context.Context) string
}

type StatementType string

const (
	SELECT StatementType = "SELECT"
	UPDATE StatementType = "UPDATE"
	DELETE StatementType = "DELETE"
	CREATE StatementType = "CREATE"
)
