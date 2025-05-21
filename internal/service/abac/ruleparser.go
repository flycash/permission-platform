package abac

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/pkg/checker"
	"github.com/ecodeclub/ekit/mapx"
	"github.com/gotomicro/ego/core/elog"
)

// 规则解析器
type RuleParser interface {
	// Check values 是 attr_id 到值的映射
	Check(rules []domain.PolicyRule, subject domain.ABACObject, resource domain.ABACObject, environment domain.ABACObject) bool
}

type ruleParser struct {
	checkSelector checker.AttributeCheckerSelector
	logger        *elog.Component
}

func NewRuleParser(checkBuilder checker.AttributeCheckerSelector) RuleParser {
	return &ruleParser{
		checkSelector: checkBuilder,
		logger:        elog.DefaultLogger,
	}
}

func (r *ruleParser) Check(rules []domain.PolicyRule, subject domain.ABACObject, resource domain.ABACObject, environment domain.ABACObject) bool {
	// 因为 Rule 是按照 attr_id 来设计的，所以我们可以合并一下所有的对象的属性取值
	// 它们的 attr_id 都是不同的，所以不会有问题
	values := mapx.Merge(subject.ValuesMap(), resource.ValuesMap(), environment.ValuesMap())

	res := true
	for idx := range rules {
		// 各个独立的规则之间是且的关系
		res = res && r.checkOneRule(rules[idx], values)
	}
	return res
}

func (r *ruleParser) checkOneRule(rule domain.PolicyRule, values map[int64]string) bool {
	if rule.LeftRule == nil && rule.RightRule == nil {
		attrID := rule.AttrDef.ID
		actualVal := values[attrID]
		checkor, err := r.checkSelector.Select(rule.AttrDef.DataType)
		if err != nil {
			return false
		}
		ok, err := checkor.CheckAttribute(rule.Value, actualVal, rule.Operator)
		if err != nil {
			return false
		}
		return ok
	}
	left, right := true, true
	if rule.LeftRule != nil {
		left = r.checkOneRule(*rule.LeftRule, values)
	}
	if rule.RightRule != nil {
		right = r.checkOneRule(*rule.RightRule, values)
	}
	switch rule.Operator {
	case domain.AND:
		return left && right
	case domain.OR:
		return left || right
	case domain.NOT:
		return !right
	default:
		return false
	}
}
