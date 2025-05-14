package checker

import (
	"encoding/json"
	"strconv"

	"gitee.com/flycash/permission-platform/internal/domain"
)

type FloatChecker struct{}

func (f FloatChecker) CheckAttribute(wantVal, actualVal string, op domain.RuleOperator) (bool, error) {
	if isSlice(op) {
		list, convActualVal, err := f.getSliceData(wantVal, actualVal)
		if err != nil {
			return false, err
		}
		return sliceCheck[float64](list, convActualVal, op)
	}
	convWantVal, convActualVal, err := f.getData(wantVal, actualVal)
	if err != nil {
		return false, err
	}
	return baseChecker[float64](convWantVal, convActualVal, op)
}

func (FloatChecker) getData(wantVal, actualVal string) (convWantVal, convActualVal float64, err error) {
	convWantVal, err = strconv.ParseFloat(wantVal, 64)
	if err != nil {
		return 0, 0, err
	}
	convActualVal, err = strconv.ParseFloat(actualVal, 64)
	return convWantVal, convActualVal, err
}

func (FloatChecker) getSliceData(wantVal, actualVal string) (res []float64, ans float64, err error) {
	err = json.Unmarshal([]byte(wantVal), &res)
	if err != nil {
		return nil, 0, err
	}
	convActualVal, err := strconv.ParseFloat(actualVal, 64)
	return res, convActualVal, err
}
