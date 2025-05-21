package evaluator

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
	"gitee.com/flycash/permission-platform/internal/service/abac/converter"
)

type ArrayEvaluator struct {
	converter converter.Converter[[]string]
}

func (a ArrayEvaluator) Evaluate(wantVal, actualVal string, op domain.RuleOperator) (bool, error) {
	wantValArray, err := a.converter.Decode(wantVal)
	if err != nil {
		return false, err
	}
	actualValArray, err := a.converter.Decode(actualVal)
	if err != nil {
		return false, err
	}
	if len(actualValArray) == 0 {
		return false, nil
	}
	switch op {
	case domain.AnyMatch:
		return a.checkAnyMatch(wantValArray, actualValArray), nil
	case domain.AllMatch:
		return a.checkAllMatch(wantValArray, actualValArray), nil
	default:
		return false, errs.ErrUnknownOperator
	}
}

func (a ArrayEvaluator) checkAnyMatch(wantVal, actualVal []string) bool {
	wantValMap := make(map[string]struct{})
	for idx := range wantVal {
		wantValMap[wantVal[idx]] = struct{}{}
	}
	for idx := range actualVal {
		if _, ok := wantValMap[actualVal[idx]]; ok {
			return true
		}
	}
	return false
}

func (a ArrayEvaluator) checkAllMatch(wantVal, actualVal []string) bool {
	if len(actualVal) == 0 {
		return false
	}
	wantValMap := make(map[string]struct{})
	for idx := range wantVal {
		wantValMap[wantVal[idx]] = struct{}{}
	}
	for idx := range actualVal {
		if _, ok := wantValMap[actualVal[idx]]; !ok {
			return false
		}
	}
	return true
}
