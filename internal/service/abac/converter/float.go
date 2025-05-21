package converter

import "strconv"

type FloatConverter struct{}

func NewFloatConverter() FloatConverter {
	return FloatConverter{}
}

func (FloatConverter) Decode(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

func (FloatConverter) Encode(t float64) (string, error) {
	return strconv.FormatFloat(t, 'f', -1, 64), nil
}
