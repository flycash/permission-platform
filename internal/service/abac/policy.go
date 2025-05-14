package abac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
)

type PolicySvc interface {
	Save(ctx context.Context, policy domain.Policy) (int64, error)
	Delete(ctx context.Context, bizID, id int64) error
	First(ctx context.Context, id int64) (domain.Policy, error) // 包含规则
	SaveRule(ctx context.Context, bizID, policyID int64, rule domain.PolicyRule) (int64, error)
	DeleteRule(ctx context.Context, ruleID int64) error
	FindPoliciesByPermissionIDs(ctx context.Context, bizID int64, permissionIDs []int64) ([]domain.Policy, error)
	SavePermissionPolicy(ctx context.Context, bizID, policyID, permissionID int64, effect domain.Effect) error
	FindPolicies(ctx context.Context, bizID int64, offset, limit int) (int64, []domain.Policy, error)
}

type policySvc struct {
	repository.PolicyRepo
}

func NewPolicySvc(repo repository.PolicyRepo) PolicySvc {
	return &policySvc{repo}
}
