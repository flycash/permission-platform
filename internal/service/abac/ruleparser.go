package abac

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/pkg/checker"
	"github.com/gotomicro/ego/core/elog"
)

// 规则解析器
type RuleParser interface {
	Check(req AttributeValReq, rules []*domain.PolicyRule) bool
}

type AttributeValReq struct {
	// 主体对象
	subject *domain.SubjectObject
	// 资源对象
	resource *domain.ResourceObject
	// 环境
	environment *domain.EnvironmentObject
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

func (r *ruleParser) Check(attributes AttributeValReq, rules []*domain.PolicyRule) bool {
	res := true
	for idx := range rules {
		// 各个独立的规则之间是且的关系
		res = res && r.checkOneRule(attributes, rules[idx])
	}
	return res
}

//nolint:funlen // 忽略
func (r *ruleParser) checkOneRule(attributes AttributeValReq, rule *domain.PolicyRule) bool {
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
		checker, err := r.checkSelector.Select(rule.AttributeDefinition.DataType)
		if err != nil {
			return false
		}
		ok, err := checker.CheckAttribute(rule.Value, actualVal, rule.Operator)
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
