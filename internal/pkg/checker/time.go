package checker

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ecodeclub/ekit/slice"

	"gitee.com/flycash/permission-platform/internal/domain"
)

const (
	timeType  string = "time"
	dayType          = "day"
	monthType        = "month"
	weekType         = "week"

	// Time related constants
	hoursInDay    = 24
	minutesInHour = 60
	daysInWeek    = 7
	minWeekday    = 0
	maxWeekday    = 6
	minMonthDay   = 1
	maxMonthDay   = 31
	number2       = 2
)

// timeRule represents a time-based rule
type timeRule struct {
	Type     string // "time", "day", "week", "month"
	Value    string
	Operator string
}

// parseTimeRule parses a time rule string into a timeRule struct
func parseTimeRule(rule string) (*timeRule, error) {
	rule = strings.TrimPrefix(rule, "@")
	parts := strings.SplitN(rule, "(", number2)
	if len(parts) != number2 {
		return nil, fmt.Errorf("invalid rule format: %s", rule)
	}

	ruleType := parts[0]
	if !slice.Contains[string]([]string{timeType, dayType, monthType, weekType}, ruleType) {
		return nil, fmt.Errorf("invalid rule type: %s", ruleType)
	}
	value := strings.TrimSuffix(parts[1], ")")

	return &timeRule{
		Type:  ruleType,
		Value: value,
	}, nil
}

type TimeChecker struct{}

func (t TimeChecker) CheckAttribute(wantVal, actualVal string, op domain.RuleOperator) (bool, error) {
	timestamp, err := strconv.ParseInt(actualVal, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp: %s", actualVal)
	}

	rule, err := parseTimeRule(wantVal)
	if err != nil {
		return false, err
	}

	// Convert millisecond timestamp to time.Time
	checkTime := time.UnixMilli(timestamp)

	// Check based on rule type
	switch rule.Type {
	case timeType:
		return checkExactTime(checkTime, rule, op)
	case dayType:
		return checkDailyTime(checkTime, rule, op)
	case weekType:
		return checkWeeklyTime(checkTime, rule, op)
	case monthType:
		return checkMonthlyTime(checkTime, rule, op)
	default:
		return false, fmt.Errorf("unknown rule type: %s", rule.Type)
	}
}

func checkExactTime(t time.Time, rule *timeRule, op domain.RuleOperator) (bool, error) {
	targetTime, err := strconv.ParseInt(rule.Value, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid time value: %s", rule.Value)
	}

	target := time.UnixMilli(targetTime)
	return compareTimes(t, target, op)
}

func checkDailyTime(t time.Time, rule *timeRule, op domain.RuleOperator) (bool, error) {
	timeStr := rule.Value
	hour, minute, err := parseTimeOfDay(timeStr)
	if err != nil {
		return false, err
	}

	today := time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, t.Location())
	return compareTimes(t, today, op)
}

func checkWeeklyTime(t time.Time, rule *timeRule, op domain.RuleOperator) (bool, error) {
	parts := strings.Split(rule.Value, ",")
	if len(parts) != number2 {
		return false, fmt.Errorf("invalid weekly time format: %s", rule.Value)
	}

	weekday, err := strconv.Atoi(parts[0])
	if err != nil || weekday < minWeekday || weekday > maxWeekday {
		return false, fmt.Errorf("invalid weekday: %s", parts[0])
	}

	hour, minute, err := parseTimeOfDay(parts[1])
	if err != nil {
		return false, err
	}

	// 直接设置目标星期几
	target := time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, t.Location())
	target = target.AddDate(0, 0, weekday-int(t.Weekday()))
	return compareTimes(t, target, op)
}

func checkMonthlyTime(t time.Time, rule *timeRule, op domain.RuleOperator) (bool, error) {
	parts := strings.Split(rule.Value, ",")
	if len(parts) != number2 {
		return false, fmt.Errorf("invalid monthly time format: %s", rule.Value)
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil || day < minMonthDay || day > maxMonthDay {
		return false, fmt.Errorf("invalid day of month: %s", parts[0])
	}

	hour, minute, err := parseTimeOfDay(parts[1])
	if err != nil {
		return false, err
	}

	// 只处理当前月的时间
	targetDate := time.Date(t.Year(), t.Month(), day, hour, minute, 0, 0, t.Location())
	return compareTimes(t, targetDate, op)
}

func parseTimeOfDay(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != number2 {
		return 0, 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour >= hoursInDay {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute >= minutesInHour {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}

	return hour, minute, nil
}

func compareTimes(t1, t2 time.Time, operator domain.RuleOperator) (bool, error) {
	switch operator {
	case domain.Greater:
		return t1.After(t2), nil
	case domain.Less:
		return t1.Before(t2), nil
	case domain.GreaterOrEqual:
		return !t1.Before(t2), nil
	case domain.LessOrEqual:
		return !t1.After(t2), nil
	default:
		return false, fmt.Errorf("错误的操作符: %s", operator)
	}
}
