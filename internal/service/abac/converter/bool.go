package converter

import "strconv"

type BoolConverter struct{}

func NewBoolConverter() BoolConverter {
	return BoolConverter{}
}

func (b BoolConverter) Decode(str string) (bool, error) {
	return strconv.ParseBool(str)
}

func (b BoolConverter) Encode(t bool) (string, error) {
	return strconv.FormatBool(t), nil
}
