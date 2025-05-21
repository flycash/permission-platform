package repository

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
)

func genDomainPolicyRules(rules []dao.PolicyRule) []domain.PolicyRule {
	ruleMap := make(map[int64]dao.PolicyRule)
	for idx := range rules {
		rule := rules[idx]
		ruleMap[rule.ID] = rule
	}
	rootRules := findRootRules(rules, ruleMap)
	for idx := range rootRules {
		rootRule := rootRules[idx]
		rootRules[idx] = genRule(rootRule, ruleMap)
	}
	return rootRules
}

func findRootRules(rules []dao.PolicyRule, ruleMap map[int64]dao.PolicyRule) []domain.PolicyRule {
	childMap := make(map[int64]struct{}, len(rules))
	for idx := range rules {
		rule := rules[idx]
		if rule.Left > 0 {
			childMap[rule.Left] = struct{}{}
		}
		if rule.Right > 0 {
			childMap[rule.Right] = struct{}{}
		}
	}
	res := make([]domain.PolicyRule, 0, len(ruleMap))
	for ruleID := range ruleMap {
		// 说明是根节点
		if _, ok := childMap[ruleID]; !ok {
			res = append(res, domain.PolicyRule{ID: ruleID})
		}
	}
	return res
}

func genRule(rule domain.PolicyRule, ruleMap map[int64]dao.PolicyRule) domain.PolicyRule {
	ruleDao, ok := ruleMap[rule.ID]
	if !ok {
		return domain.PolicyRule{}
	}
	rule = domain.PolicyRule{
		ID: rule.ID,
		AttrDef: domain.AttributeDefinition{
			ID: ruleDao.AttrDefID,
		},
		Value:    ruleDao.Value,
		Operator: domain.RuleOperator(ruleDao.Operator),
		Ctime:    ruleDao.Ctime,
		Utime:    ruleDao.Utime,
	}
	if ruleDao.Left > 0 {
		left := genRule(domain.PolicyRule{ID: ruleDao.Left}, ruleMap)
		rule.LeftRule = &left
	}
	if ruleDao.Right > 0 {
		right := genRule(domain.PolicyRule{ID: ruleDao.Right}, ruleMap)
		rule.RightRule = &right
	}
	return rule
}
