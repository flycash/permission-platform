package checker

import (
	"fmt"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
)

func TestTimeChecker_CheckAttribute(t *testing.T) {
	t.Parallel()
	checker := TimeChecker{}
	now := time.Now()

	tests := []struct {
		name      string
		wantVal   string
		actualVal string
		op        domain.RuleOperator
		want      bool
		wantErr   bool
	}{
		// @time tests
		{
			name:      "exact time greater than",
			wantVal:   "@time(1234567890000)",
			actualVal: "1234567891000",
			op:        domain.Greater,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "exact time less than",
			wantVal:   "@time(1234567890000)",
			actualVal: "1234567889000",
			op:        domain.Less,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "exact time equal",
			wantVal:   "@time(1234567890000)",
			actualVal: "1234567890000",
			op:        domain.GreaterOrEqual,
			want:      true,
			wantErr:   false,
		},

		// @day tests
		{
			name:      "daily time greater than",
			wantVal:   "@day(14:30)",
			actualVal: fmt.Sprintf("%d", time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, now.Location()).UnixMilli()),
			op:        domain.Greater,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "daily time less than",
			wantVal:   "@day(14:30)",
			actualVal: fmt.Sprintf("%d", time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, now.Location()).UnixMilli()),
			op:        domain.Less,
			want:      true,
			wantErr:   false,
		},

		// @week tests
		{
			name:      "weekly time greater than",
			wantVal:   "@week(1,14:30)",                                                               // Monday 14:30
			actualVal: fmt.Sprintf("%d", time.Date(2024, 3, 19, 15, 0, 0, 0, time.Local).UnixMilli()), // Monday 15:00
			op:        domain.Greater,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "weekly time less than",
			wantVal:   "@week(1,14:30)",                                                               // Monday 14:30
			actualVal: fmt.Sprintf("%d", time.Date(2024, 3, 18, 14, 0, 0, 0, time.Local).UnixMilli()), // Monday 14:00
			op:        domain.Less,
			want:      true,
			wantErr:   false,
		},

		// @month tests
		{
			name:      "monthly time greater than",
			wantVal:   "@month(15,14:30)",                                                             // 15th 14:30
			actualVal: fmt.Sprintf("%d", time.Date(2024, 3, 15, 15, 0, 0, 0, time.Local).UnixMilli()), // 15th 15:00
			op:        domain.Greater,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "monthly time less than",
			wantVal:   "@month(15,14:30)",                                                             // 15th 14:30
			actualVal: fmt.Sprintf("%d", time.Date(2024, 3, 15, 14, 0, 0, 0, time.Local).UnixMilli()), // 15th 14:00
			op:        domain.Less,
			want:      true,
			wantErr:   false,
		},

		// Error cases
		{
			name:      "invalid timestamp",
			wantVal:   "@time(1234567890000)",
			actualVal: "invalid",
			op:        domain.Greater,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "invalid time format",
			wantVal:   "@day(invalid)",
			actualVal: "1234567890000",
			op:        domain.Greater,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "invalid weekday",
			wantVal:   "@week(7,14:30)", // Invalid weekday
			actualVal: "1234567890000",
			op:        domain.Greater,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "invalid month day",
			wantVal:   "@month(32,14:30)", // Invalid month day
			actualVal: "1234567890000",
			op:        domain.Greater,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "invalid hour",
			wantVal:   "@day(24:00)", // Invalid hour
			actualVal: "1234567890000",
			op:        domain.Greater,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "invalid minute",
			wantVal:   "@day(14:60)", // Invalid minute
			actualVal: "1234567890000",
			op:        domain.Greater,
			want:      false,
			wantErr:   true,
		},
	}

	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := checker.CheckAttribute(tt.wantVal, tt.actualVal, tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("TimeChecker.CheckAttribute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("TimeChecker.CheckAttribute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimeRule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		rule    string
		want    *timeRule
		wantErr bool
	}{
		{
			name: "valid time rule",
			rule: "@time(1234567890000)",
			want: &timeRule{
				Type:  timeType,
				Value: "1234567890000",
			},
			wantErr: false,
		},
		{
			name: "valid day rule",
			rule: "@day(14:30)",
			want: &timeRule{
				Type:  dayType,
				Value: "14:30",
			},
			wantErr: false,
		},
		{
			name: "valid week rule",
			rule: "@week(1,14:30)",
			want: &timeRule{
				Type:  weekType,
				Value: "1,14:30",
			},
			wantErr: false,
		},
		{
			name: "valid month rule",
			rule: "@month(15,14:30)",
			want: &timeRule{
				Type:  monthType,
				Value: "15,14:30",
			},
			wantErr: false,
		},
		{
			name:    "invalid format",
			rule:    "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid type",
			rule:    "@invalid(123)",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseTimeRule(tt.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Type != tt.want.Type {
				t.Errorf("parseTimeRule() Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Value != tt.want.Value {
				t.Errorf("parseTimeRule() Value = %v, want %v", got.Value, tt.want.Value)
			}
		})
	}
}

func TestCompareTimes(t *testing.T) {
	t.Parallel()
	t1 := time.Date(2024, 3, 15, 14, 30, 0, 0, time.Local)
	t2 := time.Date(2024, 3, 15, 15, 30, 0, 0, time.Local)

	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		operator domain.RuleOperator
		want     bool
		wantErr  bool
	}{
		{
			name:     "greater than",
			t1:       t2,
			t2:       t1,
			operator: domain.Greater,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "less than",
			t1:       t1,
			t2:       t2,
			operator: domain.Less,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "greater or equal",
			t1:       t1,
			t2:       t1,
			operator: domain.GreaterOrEqual,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "less or equal",
			t1:       t1,
			t2:       t1,
			operator: domain.LessOrEqual,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "invalid operator",
			t1:       t1,
			t2:       t2,
			operator: domain.RuleOperator("invalid"),
			want:     false,
			wantErr:  true,
		},
	}

	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := compareTimes(tt.t1, tt.t2, tt.operator)
			if (err != nil) != tt.wantErr {
				t.Errorf("compareTimes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("compareTimes() = %v, want %v", got, tt.want)
			}
		})
	}
}
