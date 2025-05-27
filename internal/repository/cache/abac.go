package cache

import (
	"context"
	"errors"
	"fmt"
	"gitee.com/flycash/permission-platform/internal/domain"
)

type ABACDefinitionCache interface {
	GetDefinitions(ctx context.Context,bizID int64) (domain.BizAttrDefinition, error)
	SetDefinitions(ctx context.Context, bizDef domain.BizAttrDefinition) error
}

func  DefKey(bizID int64) string {
	return fmt.Sprintf("abac:def:%d", bizID)
}
var ErrKeyNotFound = errors.New("key not found")