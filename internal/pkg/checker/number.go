package checker

import (
	"encoding/json"
	"strconv"

	"gitee.com/flycash/permission-platform/internal/domain"
)

type NumberChecker struct{}

func (n NumberChecker) CheckAttribute(wantVal, actualVal string, op domain.RuleOperator) (bool, error) {
	if isSlice(op) {
		list, convActualVal, err := n.getSliceData(wantVal, actualVal)
		if err != nil {
			return false, err
		}
		return sliceCheck[int64](list, convActualVal, op)
	}

	convWantVal, convActualVal, err := n.getData(wantVal, actualVal)
	if err != nil {
		return false, err
	}
	return baseChecker[int64](convWantVal, convActualVal, op)
}

func (NumberChecker) getSliceData(wantVal, actualVal string) (res []int64, ans int64, err error) {
	err = json.Unmarshal([]byte(wantVal), &res)
	if err != nil {
		return nil, 0, err
	}
	actualNumberVal, err := strconv.ParseInt(actualVal, 10, 64)
	return res, actualNumberVal, err
}

func (NumberChecker) getData(wantVal, actualVal string) (convWantVal, convActualVal int64, err error) {
	convWantVal, err = strconv.ParseInt(wantVal, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	convActualVal, err = strconv.ParseInt(actualVal, 10, 64)
	return convWantVal, convActualVal, err
}
