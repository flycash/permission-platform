package evaluator

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
)

type PolicyConditionEvaluator interface {
	Evaluate(wantVal, actualVal string, op domain.RuleOperator) (bool, error)
}

type Selector interface {
	Select(dataType domain.DataType) (PolicyConditionEvaluator, error)
}

type selector struct {
	checkerMap map[domain.DataType]PolicyConditionEvaluator
}

func NewSelector() Selector {
	return &selector{
		checkerMap: map[domain.DataType]PolicyConditionEvaluator{
			domain.DataTypeString:   NewStringEvaluator(),
			domain.DataTypeBoolean:  NewBoolEvaluator(),
			domain.DataTypeFloat:    NewFloatEvaluator(),
			domain.DataTypeNumber:   NewNumberEvaluator(),
			domain.DataTypeDatetime: NewTimeEvaluator(),
			domain.DataTypeArray:    NewArrayEvaluator(),
		},
	}
}

func (c *selector) Select(dataType domain.DataType) (PolicyConditionEvaluator, error) {
	evaluator, ok := c.checkerMap[dataType]
	if !ok {
		return nil, errs.ErrUnknownDataType
	}
	return evaluator, nil
}
