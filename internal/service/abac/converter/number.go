package converter

import (
	"strconv"
)

type NumberConverter struct{}

func NewNumberConverter() NumberConverter {
	return NumberConverter{}
}

func (NumberConverter) Decode(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func (NumberConverter) Encode(t int64) (string, error) {
	return strconv.Itoa(int(t)), nil
}
