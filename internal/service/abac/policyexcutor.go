package abac

import (
	"fmt"
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/abac/evaluator"
	"github.com/ecodeclub/ekit/mapx"
	"github.com/gotomicro/ego/core/elog"
)

// 规则解析器
type PolicyExecutor interface {
	// Check values 是 attr_id 到值的映射
	Check(policy domain.Policy, subject domain.ABACObject, resource domain.ABACObject, environment domain.ABACObject) bool
}

// 基于逻辑运算符的方法
type logicOperatorExecutor struct {
	selector evaluator.Selector
	logger   *elog.Component
}

func NewPolicyExecutor(checkBuilder evaluator.Selector) PolicyExecutor {
	return &logicOperatorExecutor{
		selector: checkBuilder,
		logger:   elog.DefaultLogger,
	}
}

func (r *logicOperatorExecutor) Check(policy domain.Policy, subject, resource, environment domain.ABACObject) bool {
	// 因为 Rule 是按照 attr_id 来设计的，所以我们可以合并一下所有的对象的属性取值
	// 它们的 attr_id 都是不同的，所以不会有问题
	smap := subject.ValuesMap()
	rmap := resource.ValuesMap()
	env := environment.ValuesMap()
	//values := mapx.Merge(subject.ValuesMap(), resource.ValuesMap(), environment.ValuesMap())
	values := mapx.Merge(smap, rmap, env)

	res := true
	for idx := range policy.Rules {
		rule := policy.Rules[idx]
		// 各个独立的规则之间是且的关系
		res = res && r.checkOneRule(rule, values)
	}
	return res
}

func (r *logicOperatorExecutor) checkOneRule(rule domain.PolicyRule, values map[int64]string) bool {
	if rule.LeftRule == nil && rule.RightRule == nil {
		attrID := rule.AttrDef.ID
		actualVal := values[attrID]
		checker, err := r.selector.Select(rule.AttrDef.DataType)
		if err != nil {
			fmt.Println("1111111111", rule.ID, err)
			return false
		}
		ok, err := checker.Evaluate(rule.Value, actualVal, rule.Operator)
		if err != nil {
			fmt.Println("222222222", rule.ID, err)
			return false
		}
		if !ok {
			fmt.Println("xxxxxxxxxx", rule.ID)
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
