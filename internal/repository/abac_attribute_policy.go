package repository

import (
	"context"

	"github.com/gotomicro/ego/core/elog"

	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
)

type PolicyRepo interface {
	Save(ctx context.Context, policy domain.Policy) (int64, error)
	Delete(ctx context.Context, bizID, id int64) error
	First(ctx context.Context, bizID, id int64) (domain.Policy, error) // 包含规则
	SaveRule(ctx context.Context, bizID, policyID int64, rule domain.PolicyRule) (int64, error)
	DeleteRule(ctx context.Context, bizID, ruleID int64) error
	FindPoliciesByPermissionIDs(ctx context.Context, bizID int64, permissionIDs []int64) ([]domain.Policy, error)
	SavePermissionPolicy(ctx context.Context, bizID, policyID, permissionID int64, effect domain.Effect) error
	FindPolicies(ctx context.Context, bizID int64, offset, limit int) (int64, []domain.Policy, error)
}

type policyRepo struct {
	policyDAO dao.PolicyDAO
	logger    *elog.Component
}

func (p *policyRepo) FindPolicies(ctx context.Context, bizID int64, offset, limit int) (int64, []domain.Policy, error) {
	var (
		count int64
		res   []domain.Policy
		eg    errgroup.Group
	)
	eg.Go(func() error {
		var eerr error
		count, eerr = p.policyDAO.PolicyListCount(ctx, bizID)
		return eerr
	})
	eg.Go(func() error {
		list, err := p.policyDAO.PolicyList(ctx, bizID, offset, limit)
		if err != nil {
			return err
		}
		res = slice.Map(list, func(_ int, src dao.Policy) domain.Policy {
			return p.toPolicyDomain(src, []dao.PolicyRule{}, map[int64]dao.PermissionPolicy{})
		})
		return nil
	})
	err := eg.Wait()
	return count, res, err
}

func (p *policyRepo) SavePermissionPolicy(ctx context.Context, bizID, policyID, permissionID int64, effect domain.Effect) error {
	err := p.policyDAO.SavePermissionPolicy(ctx, dao.PermissionPolicy{
		BizID:        bizID,
		PolicyID:     policyID,
		Effect:       string(effect),
		PermissionID: permissionID,
	})
	if err != nil {
		p.logger.Error("添加权限策略关联失败",
			elog.FieldErr(err),
			elog.Int64("bizId", bizID),
			elog.Any("policyID", policyID),
			elog.Any("permissionID", permissionID),
			elog.Any("effect", effect))
	} else {
		p.logger.Info("添加权限策略关联",
			elog.Int64("bizId", bizID),
			elog.Any("policyID", policyID),
			elog.Any("permissionID", permissionID),
			elog.Any("effect", effect))
	}
	return err
}

func NewPolicyRepository(policyDAO dao.PolicyDAO) PolicyRepo {
	return &policyRepo{
		policyDAO: policyDAO,
	}
}

func (p *policyRepo) Save(ctx context.Context, policy domain.Policy) (int64, error) {
	// 转换为 DAO 层的 Policy 对象
	policyDAO := dao.Policy{
		ID:          policy.ID,
		BizID:       policy.BizID,
		Name:        policy.Name,
		ExecuteType: string(policy.ExecuteType),
		Description: policy.Description,
		Status:      string(policy.Status),
	}
	// 保存策略
	id, err := p.policyDAO.SavePolicy(ctx, policyDAO)
	if err != nil {
		p.logger.Error("添加策略失败",
			elog.FieldErr(err),
			elog.Int64("bizId", policy.BizID),
			elog.Any("policy", policy))
	} else {
		p.logger.Info("添加策略",
			elog.Int64("bizId", policy.BizID),
			elog.Any("policy", policy))
	}
	return id, err
}

func (p *policyRepo) Delete(ctx context.Context, bizID, id int64) error {
	// 删除策略及其关联数据
	err := p.policyDAO.DeletePolicy(ctx, bizID, id)
	if err != nil {
		p.logger.Error("删除策略失败",
			elog.FieldErr(err),
			elog.Int64("bizId", bizID),
			elog.Int64("policyId", id))
	} else {
		p.logger.Info("删除策略",
			elog.Int64("bizId", bizID),
			elog.Int64("policyId", id))
	}
	return err
}

func (p *policyRepo) First(ctx context.Context, bizID, id int64) (domain.Policy, error) {
	// 获取策略基本信息
	var policy dao.Policy
	var rules []dao.PolicyRule
	var eg errgroup.Group
	eg.Go(func() error {
		var eerr error
		policy, eerr = p.policyDAO.FindPolicy(ctx, bizID, id)
		return eerr
	})

	eg.Go(func() error {
		var eerr error
		rules, eerr = p.policyDAO.FindPolicyRulesByPolicyID(ctx, bizID, id)
		return eerr
	})

	if err := eg.Wait(); err != nil {
		return domain.Policy{}, err
	}
	return p.toPolicyDomain(policy, rules, map[int64]dao.PermissionPolicy{}), nil
}

func (p *policyRepo) SaveRule(ctx context.Context, bizID, policyID int64, rule domain.PolicyRule) (int64, error) {
	// 转换为 DAO 层的规则对象
	ruleDAO := dao.PolicyRule{
		ID:        rule.ID,
		BizID:     bizID,
		PolicyID:  policyID,
		AttrDefID: rule.AttrDef.ID,
		Value:     rule.Value,
		Operator:  string(rule.Operator),
	}
	if rule.LeftRule != nil {
		ruleDAO.Left = rule.LeftRule.ID
	}
	if rule.RightRule != nil {
		ruleDAO.Right = rule.RightRule.ID
	}
	id, err := p.policyDAO.SavePolicyRule(ctx, ruleDAO)
	if err != nil {
		p.logger.Error("保存策略规则失败",
			elog.FieldErr(err),
			elog.Int64("bizId", bizID),
			elog.Int64("policyId", policyID),
			elog.Any("rule", rule),
		)
	} else {
		p.logger.Info("保存策略规则",
			elog.Int64("bizId", bizID),
			elog.Int64("policyId", policyID),
			elog.Any("rule", rule),
		)
	}
	return id, err
}

func (p *policyRepo) DeleteRule(ctx context.Context, bizID, ruleID int64) error {
	err := p.policyDAO.DeletePolicyRule(ctx, bizID, ruleID)
	if err != nil {
		p.logger.Error("删除策略规则失败",
			elog.FieldErr(err),
			elog.Int64("bizId", bizID),
			elog.Any("ruleID", ruleID),
		)
	} else {
		p.logger.Info("删除策略规则",
			elog.Int64("bizId", bizID),
			elog.Any("ruleID", ruleID),
		)
	}
	return err
}

func (p *policyRepo) FindPoliciesByPermissionIDs(ctx context.Context, bizID int64, permissionID []int64) ([]domain.Policy, error) {
	policyPermissions, err := p.policyDAO.FindPoliciesByPermission(ctx, bizID, permissionID)
	if err != nil {
		return nil, err
	}
	policyIds := slice.Map(policyPermissions, func(_ int, src dao.PermissionPolicy) int64 {
		return src.PolicyID
	})
	permissionPolicyMap := make(map[int64]dao.PermissionPolicy, len(policyIds))
	for idx := range policyPermissions {
		policyPermission := policyPermissions[idx]
		permissionPolicyMap[policyPermission.PolicyID] = policyPermission
	}
	var eg errgroup.Group
	var policies []dao.Policy
	var rules map[int64][]dao.PolicyRule
	eg.Go(func() error {
		var eerr error
		policies, eerr = p.policyDAO.FindPoliciesByIDs(ctx, policyIds)
		return eerr
	})
	eg.Go(func() error {
		var eerr error
		rules, eerr = p.policyDAO.FindPolicyRulesByPolicyIDs(ctx, policyIds)
		return eerr
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	res := make([]domain.Policy, 0, len(policies))
	for idx := range policies {
		res = append(res, p.toPolicyDomain(policies[idx], rules[policies[idx].ID], permissionPolicyMap))
	}
	return res, nil
}

func (p *policyRepo) toPolicyDomain(policy dao.Policy, rules []dao.PolicyRule, permissionPolicyMap map[int64]dao.PermissionPolicy) domain.Policy {
	domainPolicy := domain.Policy{
		ID:          policy.ID,
		BizID:       policy.BizID,
		Name:        policy.Name,
		ExecuteType: domain.ExecuteType(policy.ExecuteType),
		Description: policy.Description,
		Status:      domain.PolicyStatus(policy.Status),
		Rules:       genDomainPolicyRules(rules),
	}
	if permissionPolicy, ok := permissionPolicyMap[policy.ID]; ok {
		domainPolicy.Effect = domain.Effect(permissionPolicy.Effect)
	}
	return domainPolicy
}
