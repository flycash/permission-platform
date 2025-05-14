package checker

import (
	"strconv"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
)

type BoolChecker struct{}

func (b BoolChecker) CheckAttribute(wantVal, actualVal string, op domain.RuleOperator) (bool, error) {
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

func (b BoolChecker) getData(wantVal, actualVal string) (convWantVal, convActualVal bool, err error) {
	convWantVal, err = strconv.ParseBool(wantVal)
	if err != nil {
		return false, false, err
	}
	convActualVal, err = strconv.ParseBool(actualVal)
	return convWantVal, convActualVal, err
}
