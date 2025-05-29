package cache

import (
	"context"
	"errors"
	"fmt"

	"gitee.com/flycash/permission-platform/internal/domain"
)

type ABACDefinitionCache interface {
	GetDefinitions(ctx context.Context, bizID int64) (domain.BizAttrDefinition, error)
	SetDefinitions(ctx context.Context, bizDef domain.BizAttrDefinition) error
}

type ABACPolicyCache interface {
	GetPolicies(ctx context.Context, bizID int64, policyIds []int64) (map[int64]domain.Policy, error)
	SetPolicy(ctx context.Context, bizID int64, policies []domain.Policy) error
	DelPolicy(ctx context.Context, bizID int64, permissionID int64) error

	GetPermissionPolicy(ctx context.Context, bizID int64, permissionIDs []int64) ([]domain.Policy, error)
	SetPermissionPolicy(ctx context.Context, bizID int64, reqs []SetPermissionPolicyReq) error
}

type SetPermissionPolicyReq struct {
	PermissionID int64
	Policies     []domain.Policy
}

func DefKey(bizID int64) string {
	return fmt.Sprintf("abac:def:%d", bizID)
}

var ErrKeyNotFound = errors.New("key not found")
