package evaluator

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
	"gitee.com/flycash/permission-platform/internal/service/abac/converter"
)

type BoolEvaluator struct {
	converter converter.Converter[bool]
}

func NewBoolEvaluator() BoolEvaluator {
	return BoolEvaluator{converter.NewBoolConverter()}
}

func (b BoolEvaluator) Evaluate(wantVal, actualVal string, op domain.RuleOperator) (bool, error) {
	boolWantVal, boolActualVal, err := b.getData(wantVal, actualVal)
	if err != nil {
		return false, err
	}
	switch op {
	case domain.Equals:
		return boolWantVal == boolActualVal, nil
	case domain.NotEquals:
		return !boolWantVal == boolActualVal, nil
	default:
		return false, errs.ErrUnknownOperator
	}
}

func (b BoolEvaluator) getData(wantVal, actualVal string) (convWantVal, convActualVal bool, err error) {
	convWantVal, err = b.converter.Decode(wantVal)
	if err != nil {
		return false, false, err
	}
	convActualVal, err = b.converter.Decode(actualVal)
	return convWantVal, convActualVal, err
}
