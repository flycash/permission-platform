package abac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
)

type PermissionSvc interface {
	Check(ctx context.Context, bizID, uid int64, resource domain.Resource, actions []string, attrs domain.Attributes) (bool, error)
}

type permissionSvc struct {
	permissionRepo repository.PermissionRepository
	resourceRepo   repository.ResourceRepository
	policyRepo     repository.PolicyRepo
	valRepo        repository.AttributeValueRepository
	definitionRepo repository.AttributeDefinitionRepository
	parser         PolicyExecutor
}

func NewPermissionSvc(permissionRepo repository.PermissionRepository,
	resourceRepo repository.ResourceRepository,
	policyRepo repository.PolicyRepo,
	valRepo repository.AttributeValueRepository,
	definitionRepo repository.AttributeDefinitionRepository,
	parser PolicyExecutor,
) PermissionSvc {
	return &permissionSvc{
		permissionRepo: permissionRepo,
		resourceRepo:   resourceRepo,
		policyRepo:     policyRepo,
		valRepo:        valRepo,
		definitionRepo: definitionRepo,
		parser:         parser,
	}
}

func (p *permissionSvc) Check(ctx context.Context, bizID, uid int64, resource domain.Resource, actions []string, attrs domain.Attributes) (bool, error) {
	permissions, res, err := p.getPermissionAndRes(ctx, bizID, resource, actions)
	if err != nil {
		return false, err
	}
	permissionIds := slice.Map(permissions, func(_ int, src domain.Permission) int64 {
		return src.ID
	})
	resource.ID = res.ID
	var eg errgroup.Group
	var (
		subObj        domain.ABACObject
		resObj        domain.ABACObject
		envObj        domain.ABACObject
		policies      []domain.Policy
		bizDefinition domain.BizAttrDefinition
	)

	eg.Go(func() error {
		var eerr error
		subObj, eerr = p.valRepo.FindSubjectValue(ctx, bizID, uid)
		return eerr
	})
	eg.Go(func() error {
		var eerr error
		resObj, eerr = p.valRepo.FindResourceValue(ctx, bizID, resource.ID)
		return eerr
	})
	eg.Go(func() error {
		var eerr error
		envObj, eerr = p.valRepo.FindEnvironmentValue(ctx, bizID)
		return eerr
	})

	eg.Go(func() error {
		var eerr error
		policies, eerr = p.policyRepo.FindPoliciesByPermissionIDs(ctx, bizID, permissionIds)
		return eerr
	})

	eg.Go(func() error {
		var eerr error
		bizDefinition, eerr = p.definitionRepo.Find(ctx, bizID)
		return eerr
	})

	err = eg.Wait()
	if err != nil {
		return false, err
	}

	// 将预存属性和实时属性合并在一起，实时属性的优先级更加高
	subObj.MergeRealTimeAttrs(bizDefinition.SubjectAttrDefs, attrs.Subject)
	resObj.MergeRealTimeAttrs(bizDefinition.ResourceAttrDefs, attrs.Resource)
	envObj.MergeRealTimeAttrs(bizDefinition.EnvironmentAttrDefs, attrs.Environment)

	var hasPermit bool
	var hasDeny bool
	if len(policies) == 0 {
		return true, nil
	}
	for idx := range policies {
		policy := policies[idx]
		if p.parser.Check(policy, subObj, resObj, envObj) {
			if policy.Effect == domain.EffectAllow {
				hasPermit = true
			}
			if policy.Effect == domain.EffectDeny {
				hasDeny = true
			}
		}
	}
	// 拒绝是高优先级
	if hasDeny {
		return false, nil
	}
	if hasPermit {
		return true, nil
	}
	// 一条都没符合就返回没通过校验
	return false, nil
}

func (p *permissionSvc) getPermissionAndRes(ctx context.Context, bizID int64, resource domain.Resource, actions []string) ([]domain.Permission, domain.Resource, error) {
	var (
		eg          errgroup.Group
		permissions []domain.Permission
		res         domain.Resource
	)
	eg.Go(func() error {
		var eerr error
		permissions, eerr = p.permissionRepo.FindPermissions(ctx, bizID, resource.Type, resource.Key, actions)
		return eerr
	})
	eg.Go(func() error {
		var eerr error
		res, eerr = p.resourceRepo.FindByBizIDAndTypeAndKey(ctx, bizID, resource.Type, resource.Key)
		return eerr
	})
	err := eg.Wait()
	return permissions, res, err
}
