package checker

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
)

type Numbered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

func baseChecker[T Numbered](wantVal, actualVal T, op domain.RuleOperator) (bool, error) {
	switch op {
	case domain.Equals:
		return wantVal == actualVal, nil
	case domain.NotEquals:
		return wantVal != actualVal, nil
	case domain.Greater:
		return actualVal > wantVal, nil
	case domain.GreaterOrEqual:
		return actualVal >= wantVal, nil
	case domain.LessOrEqual:
		return actualVal <= wantVal, nil
	case domain.Less:
		return actualVal < wantVal, nil
	default:
		return false, errs.ErrUnknownOperator
	}
}

func sliceCheck[T comparable](wantVal []T, actualVal T, op domain.RuleOperator) (bool, error) {
	switch op {
	case domain.IN:
		for id := range wantVal {
			if wantVal[id] == actualVal {
				return true, nil
			}
		}
		return false, nil
	case domain.NotIn:
		for id := range wantVal {
			if wantVal[id] == actualVal {
				return false, nil
			}
		}
		return true, nil
	default:
		return false, errs.ErrUnknownOperator
	}
}

func isSlice(op domain.RuleOperator) bool {
	return op == domain.IN || op == domain.NotIn
}
