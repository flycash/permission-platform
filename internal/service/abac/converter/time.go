package converter

import (
	"strconv"
	"time"
)

type TimeConverter struct{}

func NewTimeConverter() TimeConverter {
	return TimeConverter{}
}

func (t TimeConverter) Decode(str string) (time.Time, error) {
	stamp, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.UnixMilli(stamp), nil
}

func (t TimeConverter) Encode(v time.Time) (string, error) {
	stamp := v.UnixMilli()
	return strconv.FormatInt(stamp, 10), nil
}
