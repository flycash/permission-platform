package checker

import (
	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/errs"
)

type AttributeChecker interface {
	CheckAttribute(wantVal, actualVal string, op domain.RuleOperator) (bool, error)
}

type AttributeCheckerSelector interface {
	Select(dataType domain.DataType) (AttributeChecker, error)
}

type CheckerBuilder struct {
	checkerMap map[domain.DataType]AttributeChecker
}

func NewCheckerBuilder() AttributeCheckerSelector {
	return &CheckerBuilder{
		checkerMap: map[domain.DataType]AttributeChecker{
			domain.DataTypeString:   StringChecker{},
			domain.DataTypeBoolean:  BoolChecker{},
			domain.DataTypeFloat:    FloatChecker{},
			domain.DataTypeNumber:   NumberChecker{},
			domain.DataTypeDatetime: TimeChecker{},
		},
	}
}

func (c *CheckerBuilder) Select(dataType domain.DataType) (AttributeChecker, error) {
	checker, ok := c.checkerMap[dataType]
	if !ok {
		return nil, errs.ErrUnknownDataType
	}
	return checker, nil
}
