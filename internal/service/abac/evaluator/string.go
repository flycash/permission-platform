package evaluator

import (
	"encoding/json"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
)

type StringEvaluator struct{}

func (s StringEvaluator) Evaluate(wantVal, actualVal string, op domain.RuleOperator) (bool, error) {
	if isSlice(op) {
		list, convActualVal, err := s.getSliceData(wantVal, actualVal)
		if err != nil {
			return false, err
		}
		return sliceEvaluator[string](list, convActualVal, op)
	}
	switch op {
	case domain.Equals:
		return wantVal == actualVal, nil
	case domain.NotEquals:
		return wantVal != actualVal, nil
	default:
		return false, errs.ErrUnknownOperator
	}
}

func (StringEvaluator) getSliceData(wantVal, actualVal string) (res []string, ans string, err error) {
	err = json.Unmarshal([]byte(wantVal), &res)
	if err != nil {
		return nil, "", err
	}
	return res, actualVal, err
}
