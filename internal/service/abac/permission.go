package abac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
)

type PermissionSvc interface {
	Check(ctx context.Context, bizID, uid int64,
		permission domain.Permission, attrs domain.PermissionRequest) (bool, error)
}

type permissionSvc struct {
	permissionRepo repository.RBACRepository
	policyRepo     repository.PolicyRepo
	valRepo        repository.AttributeValueRepository
	definitionRepo repository.AttributeDefinitionRepository
	parser         RuleParser
}

func NewPermissionSvc(permissionRepo repository.RBACRepository,
	policyRepo repository.PolicyRepo,
	valRepo repository.AttributeValueRepository,
	definitionRepo repository.AttributeDefinitionRepository,
	parser RuleParser,
) PermissionSvc {
	return &permissionSvc{
		permissionRepo: permissionRepo,
		policyRepo:     policyRepo,
		valRepo:        valRepo,
		definitionRepo: definitionRepo,
		parser:         parser,
	}
}

func (p *permissionSvc) Check(ctx context.Context, bizID, uid int64,
	permission domain.Permission, attrs domain.PermissionRequest,
) (bool, error) {
	permissions, res, err := p.getPermissionAndRes(ctx, bizID, permission)
	if err != nil {
		return false, err
	}
	permissionIds := slice.Map(permissions, func(_ int, src domain.Permission) int64 {
		return src.ID
	})
	permission.ResourceID = res.ID
	var eg errgroup.Group
	var (
		subObj        domain.SubjectObject
		resObj        domain.ResourceObject
		envObj        domain.EnvironmentObject
		policies      []domain.Policy
		bizDefinition domain.BizDefinition
	)
	eg.Go(func() error {
		var eerr error
		subObj, resObj, envObj, eerr = p.getAttributesVal(ctx, bizID, uid, permission)
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

	attributeValReq := p.buildAttributeValReq(bizDefinition, &subObj, &resObj, &envObj, attrs)
	var hasPermit bool
	var hasDeny bool
	for idx := range policies {
		policy := policies[idx]
		p.setPolicyDefinition(bizDefinition, policy.Rules)
		if p.parser.Check(attributeValReq, policy.Rules) {
			if policy.Effect == domain.EffectPermit {
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

func (p *permissionSvc) getPermissionAndRes(ctx context.Context, bizID int64, permission domain.Permission) ([]domain.Permission, domain.Resource, error) {
	var (
		eg          errgroup.Group
		permissions []domain.Permission
		res         domain.Resource
	)
	eg.Go(func() error {
		var eerr error
		permissions, eerr = p.permissionRepo.FindPermissions(ctx, bizID, permission.ResourceKey, permission.Action)
		return eerr
	})
	eg.Go(func() error {
		var eerr error
		res, eerr = p.permissionRepo.FindResource(ctx, bizID, permission.ResourceKey)
		return eerr
	})
	err := eg.Wait()
	return permissions, res, err
}

func (p *permissionSvc) buildAttributeValReq(bizDefinition domain.BizDefinition,
	subObj *domain.SubjectObject,
	resObj *domain.ResourceObject,
	envObj *domain.EnvironmentObject,
	attrs domain.PermissionRequest,
) AttributeValReq {
	p.setObjDefinition(bizDefinition, subObj, resObj, envObj)
	for name, value := range attrs.EnvironmentAttrs {
		definition, ok := bizDefinition.EnvironmentAttrs.GetDefinitionWithName(name)
		if !ok {
			continue
		}
		envObj.SetAttributeVal(value, definition)
	}
	for name, value := range attrs.SubjectAttrs {
		definition, ok := bizDefinition.SubjectAttrs.GetDefinitionWithName(name)
		if !ok {
			continue
		}
		subObj.SetAttributeVal(value, definition)
	}
	for name, value := range attrs.ResourceAttrs {
		definition, ok := bizDefinition.ResourceAttrs.GetDefinitionWithName(name)
		if !ok {
			continue
		}
		resObj.SetAttributeVal(value, definition)
	}
	return AttributeValReq{
		subject:     subObj,
		resource:    resObj,
		environment: envObj,
	}
}

func (p *permissionSvc) getAttributesVal(ctx context.Context, bizID, uid int64, permission domain.Permission) (subObj domain.SubjectObject, resObj domain.ResourceObject, envObj domain.EnvironmentObject, err error) {
	var eg errgroup.Group
	eg.Go(func() error {
		var eerr error
		subObj, eerr = p.valRepo.FindSubjectValue(ctx, bizID, uid)
		return eerr
	})
	eg.Go(func() error {
		var eerr error
		resObj, eerr = p.valRepo.FindResourceValue(ctx, bizID, permission.ResourceID)
		return eerr
	})
	eg.Go(func() error {
		var eerr error
		envObj, eerr = p.valRepo.FindEnvironmentValue(ctx, bizID)
		return eerr
	})
	err = eg.Wait()
	return
}

func (p *permissionSvc) setObjDefinition(bizDefinition domain.BizDefinition,
	subObj *domain.SubjectObject,
	resObj *domain.ResourceObject,
	envObj *domain.EnvironmentObject,
) {
	for idx := range subObj.AttributeValues {
		val := subObj.AttributeValues[idx]
		subObj.AttributeValues[idx].Definition, _ = bizDefinition.SubjectAttrs.GetDefinition(val.Definition.ID)
	}
	for idx := range resObj.AttributeValues {
		val := resObj.AttributeValues[idx]
		resObj.AttributeValues[idx].Definition, _ = bizDefinition.ResourceAttrs.GetDefinition(val.Definition.ID)
	}
	for idx := range envObj.AttributeValues {
		val := envObj.AttributeValues[idx]
		envObj.AttributeValues[idx].Definition, _ = bizDefinition.EnvironmentAttrs.GetDefinition(val.Definition.ID)
	}
}

func (p *permissionSvc) setPolicyDefinition(
	bizDefinition domain.BizDefinition,
	rules []*domain.PolicyRule,
) {
	for idx := range rules {
		rule := rules[idx]
		p.setPolicyRuleDefinition(bizDefinition, rule)
	}
}

func (p *permissionSvc) setPolicyRuleDefinition(
	bizDefinition domain.BizDefinition,
	rule *domain.PolicyRule,
) {
	aid := rule.AttributeDefinition.ID
	if v, ok := bizDefinition.SubjectAttrs.GetDefinition(aid); ok {
		rule.AttributeDefinition = v
	}
	if v, ok := bizDefinition.ResourceAttrs.GetDefinition(aid); ok {
		rule.AttributeDefinition = v
	}
	if v, ok := bizDefinition.EnvironmentAttrs.GetDefinition(aid); ok {
		rule.AttributeDefinition = v
	}
	if rule.LeftRule != nil {
		p.setPolicyRuleDefinition(bizDefinition, rule.LeftRule)
	}
	if rule.RightRule != nil {
		p.setPolicyRuleDefinition(bizDefinition, rule.RightRule)
	}
}
