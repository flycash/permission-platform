package evaluator

import (
	"encoding/json"
	"strconv"

	"gitee.com/flycash/permission-platform/internal/service/abac/converter"

	"gitee.com/flycash/permission-platform/internal/domain"
)

type FloatEvaluator struct {
	converter converter.FloatConverter
}

func NewFloatEvaluator() FloatEvaluator {
	return FloatEvaluator{
		converter: converter.NewFloatConverter(),
	}
}

func (f FloatEvaluator) Evaluate(wantVal, actualVal string, op domain.RuleOperator) (bool, error) {
	if isSlice(op) {
		list, convActualVal, err := f.getSliceData(wantVal, actualVal)
		if err != nil {
			return false, err
		}
		return sliceEvaluator[float64](list, convActualVal, op)
	}
	convWantVal, convActualVal, err := f.getData(wantVal, actualVal)
	if err != nil {
		return false, err
	}
	return baseEvaluator[float64](convWantVal, convActualVal, op)
}

func (f FloatEvaluator) getData(wantVal, actualVal string) (convWantVal, convActualVal float64, err error) {
	convWantVal, err = f.converter.Decode(wantVal)
	if err != nil {
		return 0, 0, err
	}
	convActualVal, err = converter.NewFloatConverter().Decode(actualVal)
	return convWantVal, convActualVal, err
}

func (FloatEvaluator) getSliceData(wantVal, actualVal string) (res []float64, ans float64, err error) {
	err = json.Unmarshal([]byte(wantVal), &res)
	if err != nil {
		return nil, 0, err
	}
	convActualVal, err := strconv.ParseFloat(actualVal, 64)
	return res, convActualVal, err
}
