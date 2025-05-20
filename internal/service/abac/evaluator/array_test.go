package evaluator

import (
	"testing"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/abac/converter"
)

func TestArrayEvaluator_Evaluate(t *testing.T) {
	t.Parallel()
	evaluator := ArrayEvaluator{
		converter: converter.NewArrayConverter(),
	}

	tests := []struct {
		name      string
		wantVal   string
		actualVal string
		op        domain.RuleOperator
		want      bool
		wantErr   bool
	}{
		{
			name:      "any match success",
			wantVal:   `["a", "b", "c"]`,
			actualVal: `["d", "b", "e"]`,
			op:        domain.AnyMatch,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "any match failure",
			wantVal:   `["a", "b", "c"]`,
			actualVal: `["d", "e", "f"]`,
			op:        domain.AnyMatch,
			want:      false,
			wantErr:   false,
		},
		{
			name:      "all match success",
			wantVal:   `["a", "b", "c"]`,
			actualVal: `["a", "b", "c", "d"]`,
			op:        domain.AllMatch,
			want:      false,
			wantErr:   false,
		},
		{
			name:      "all match failure",
			wantVal:   `["a", "b", "c"]`,
			actualVal: `["a", "b", "d"]`,
			op:        domain.AllMatch,
			want:      false,
			wantErr:   false,
		},
		{
			name:      "invalid json want value",
			wantVal:   `["a", "b", "c"`, // 缺少右括号
			actualVal: `["a", "b", "c"]`,
			op:        domain.AnyMatch,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "invalid json actual value",
			wantVal:   `["a", "b", "c"]`,
			actualVal: `["a", "b", "c"`, // 缺少右括号
			op:        domain.AnyMatch,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "unknown operator",
			wantVal:   `["a", "b", "c"]`,
			actualVal: `["a", "b", "c"]`,
			op:        domain.Equals, // 不支持的操作符
			want:      false,
			wantErr:   true,
		},
	}

	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := evaluator.Evaluate(tt.wantVal, tt.actualVal, tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArrayEvaluator.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ArrayEvaluator.Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayEvaluator_checkAnyMatch(t *testing.T) {
	t.Parallel()
	evaluator := ArrayEvaluator{}

	tests := []struct {
		name      string
		wantVal   []string
		actualVal []string
		want      bool
	}{
		{
			name:      "match found",
			wantVal:   []string{"a", "b", "c"},
			actualVal: []string{"d", "b", "e"},
			want:      true,
		},
		{
			name:      "no match",
			wantVal:   []string{"a", "b", "c"},
			actualVal: []string{"d", "e", "f"},
			want:      false,
		},
		{
			name:      "empty arrays",
			wantVal:   []string{},
			actualVal: []string{},
			want:      false,
		},
		{
			name:      "empty want array",
			wantVal:   []string{},
			actualVal: []string{"a", "b", "c"},
			want:      false,
		},
		{
			name:      "empty actual array",
			wantVal:   []string{"a", "b", "c"},
			actualVal: []string{},
			want:      false,
		},
	}

	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := evaluator.checkAnyMatch(tt.wantVal, tt.actualVal)
			if got != tt.want {
				t.Errorf("ArrayEvaluator.checkAnyMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayEvaluator_checkAllMatch(t *testing.T) {
	t.Parallel()
	evaluator := ArrayEvaluator{}

	tests := []struct {
		name      string
		wantVal   []string
		actualVal []string
		want      bool
	}{
		{
			name:      "all elements match",
			wantVal:   []string{"a", "b", "c"},
			actualVal: []string{"a", "b", "c", "d"},
			want:      false,
		},
		{
			name:      "not all elements match",
			wantVal:   []string{"a", "b", "c"},
			actualVal: []string{"a", "b", "d"},
			want:      false,
		},
		{
			name:      "empty arrays",
			wantVal:   []string{},
			actualVal: []string{},
			want:      false,
		},
		{
			name:      "empty want array",
			wantVal:   []string{},
			actualVal: []string{"a", "b", "c"},
			want:      false,
		},
		{
			name:      "empty actual array",
			wantVal:   []string{"a", "b", "c"},
			actualVal: []string{},
			want:      false,
		},
	}

	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := evaluator.checkAllMatch(tt.wantVal, tt.actualVal)
			if got != tt.want {
				t.Errorf("ArrayEvaluator.checkAllMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
