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
	GetPolicies(ctx context.Context, bizID int64) ([]domain.Policy, error)
	SetPolicy(ctx context.Context, bizID int64, policies []domain.Policy) error
	DelPolicy(ctx context.Context, bizID int64, policyID int64) error
}

type ABACAttributeValCache interface {
	GetAttrResObj(ctx context.Context, bizID int64, ids int64) (domain.ABACObject, error)
	SetAttrResObj(ctx context.Context,objs []domain.ABACObject) error
	GetAttrSubObj(ctx context.Context, bizID int64, ids int64) (domain.ABACObject, error)
	SetAttrSubObj(ctx context.Context,objs []domain.ABACObject) error
	GetAttrEnvObj(ctx context.Context, bizID int64) (domain.ABACObject, error)
	SetAttrEnvObj(ctx context.Context,objs []domain.ABACObject) error
}

type SetPermissionPolicyReq struct {
	PermissionID int64
	Policies     []domain.Policy
}

func DefKey(bizID int64) string {
	return fmt.Sprintf("abac:def:%d", bizID)
}

var ErrKeyNotFound = errors.New("key not found")
