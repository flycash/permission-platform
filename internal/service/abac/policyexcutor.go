package abac

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/abac/evaluator"
	"github.com/gotomicro/ego/core/elog"
)

// 策略执行器
type PolicyExecutor interface {
	Check(req AttributeValReq, policy domain.Policy) bool
}

type AttributeValReq struct {
	// 主体对象
	subject *domain.SubjectObject
	// 资源对象
	resource *domain.ResourceObject
	// 环境
	environment *domain.EnvironmentObject
}

// 基于逻辑运算符的方法
type logicOperatorExecutor struct {
	selector evaluator.Selector
	logger   *elog.Component
}

func NewRuleParser(checkBuilder evaluator.Selector) PolicyExecutor {
	return &logicOperatorExecutor{
		selector: checkBuilder,
		logger:   elog.DefaultLogger,
	}
}

func (r *logicOperatorExecutor) Check(attributes AttributeValReq, policy domain.Policy) bool {
	res := true
	for idx := range policy.Rules {
		rule := policy.Rules[idx]
		// 各个独立的规则之间是且的关系
		res = res && r.checkOneRule(attributes, rule)
	}
	return res
}

//nolint:funlen // 忽略
func (r *logicOperatorExecutor) checkOneRule(attributes AttributeValReq, rule *domain.PolicyRule) bool {
	if rule.LeftRule == nil && rule.RightRule == nil {
		var actualVal string
		switch rule.AttributeDefinition.EntityType {
		case domain.EntityTypeSubject:
			wantVal, err := attributes.subject.AttributeVal(rule.AttributeDefinition.ID)
			if err != nil {
				return false
			}
			actualVal = wantVal.Value
		case domain.EntityTypeResource:
			wantVal, err := attributes.resource.AttributeVal(rule.AttributeDefinition.ID)
			if err != nil {
				return false
			}
			actualVal = wantVal.Value
		case domain.EntityTypeEnvironment:
			wantVal, err := attributes.environment.AttributeVal(rule.AttributeDefinition.ID)
			if err != nil {
				return false
			}
			actualVal = wantVal.Value
		default:
			return false
		}
		eva, err := r.selector.Select(rule.AttributeDefinition.DataType)
		if err != nil {
			return false
		}
		ok, err := eva.Evaluate(rule.Value, actualVal, rule.Operator)
		if err != nil {
			return false
		}
		return ok
	}
	left, right := true, true
	if rule.LeftRule != nil {
		left = r.checkOneRule(attributes, rule.LeftRule)
	}
	if rule.RightRule != nil {
		right = r.checkOneRule(attributes, rule.RightRule)
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
