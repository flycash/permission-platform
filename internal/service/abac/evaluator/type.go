package evaluator

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
)

type PolicyRuleEvaluator interface {
	Evaluate(wantVal, actualVal string, op domain.RuleOperator) (bool, error)
}

type Selector interface {
	Select(dataType domain.DataType) (PolicyRuleEvaluator, error)
}

type selector struct {
	checkerMap map[domain.DataType]PolicyRuleEvaluator
}

func NewSelector() Selector {
	return &selector{
		checkerMap: map[domain.DataType]PolicyRuleEvaluator{
			domain.DataTypeString:   StringEvaluator{},
			domain.DataTypeBoolean:  NewBoolEvaluator(),
			domain.DataTypeFloat:    NewFloatEvaluator(),
			domain.DataTypeNumber:   NewNumberEvaluator(),
			domain.DataTypeDatetime: NewTimeEvaluator(),
		},
	}
}

func (c *selector) Select(dataType domain.DataType) (PolicyRuleEvaluator, error) {
	evaluator, ok := c.checkerMap[dataType]
	if !ok {
		return nil, errs.ErrUnknownDataType
	}
	return evaluator, nil
}
