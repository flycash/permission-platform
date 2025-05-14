package repository

import (
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/stretchr/testify/assert"
)

func TestGenDomainPolicyRules(t *testing.T) {
	t.Parallel()
	now := time.Now().UnixMilli()
	tests := []struct {
		name     string
		rules    []dao.PolicyRule
		expected []*domain.PolicyRule
	}{
		{
			name:     "empty rules",
			rules:    []dao.PolicyRule{},
			expected: []*domain.PolicyRule{},
		},
		{
			name: "single rule",
			rules: []dao.PolicyRule{
				{
					ID:          1,
					AttributeID: 100,
					Value:       "test",
					Operator:    "eq",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
			},
			expected: []*domain.PolicyRule{
				{
					ID: 1,
					AttributeDefinition: domain.AttributeDefinition{
						ID: 100,
					},
					Value:    "test",
					Operator: "eq",
					Ctime:    now,
					Utime:    now,
				},
			},
		},
		{
			name: "nested rules",
			rules: []dao.PolicyRule{
				{
					ID:          1,
					AttributeID: 100,
					Value:       "root",
					Operator:    "and",
					Left:        2,
					Right:       3,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          2,
					AttributeID: 101,
					Value:       "left",
					Operator:    "neq",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          3,
					AttributeID: 102,
					Value:       "right",
					Operator:    "gt",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
			},
			expected: []*domain.PolicyRule{
				{
					ID: 1,
					AttributeDefinition: domain.AttributeDefinition{
						ID: 100,
					},
					Value:    "root",
					Operator: "and",
					Ctime:    now,
					Utime:    now,
					LeftRule: &domain.PolicyRule{
						ID: 2,
						AttributeDefinition: domain.AttributeDefinition{
							ID: 101,
						},
						Value:    "left",
						Operator: "neq",
						Ctime:    now,
						Utime:    now,
					},
					RightRule: &domain.PolicyRule{
						ID: 3,
						AttributeDefinition: domain.AttributeDefinition{
							ID: 102,
						},
						Value:    "right",
						Operator: "gt",
						Ctime:    now,
						Utime:    now,
					},
				},
			},
		},
		{
			name: "multiple root rules",
			rules: []dao.PolicyRule{
				{
					ID:          1,
					AttributeID: 100,
					Value:       "root1",
					Operator:    "eq",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          2,
					AttributeID: 101,
					Value:       "root2",
					Operator:    "neq",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
			},
			expected: []*domain.PolicyRule{
				{
					ID: 1,
					AttributeDefinition: domain.AttributeDefinition{
						ID: 100,
					},
					Value:    "root1",
					Operator: "eq",
					Ctime:    now,
					Utime:    now,
				},
				{
					ID: 2,
					AttributeDefinition: domain.AttributeDefinition{
						ID: 101,
					},
					Value:    "root2",
					Operator: "neq",
					Ctime:    now,
					Utime:    now,
				},
			},
		},
		{
			name: "complex nested rules with multiple levels",
			rules: []dao.PolicyRule{
				{
					ID:          1,
					AttributeID: 100,
					Value:       "root",
					Operator:    "and",
					Left:        2,
					Right:       3,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          2,
					AttributeID: 101,
					Value:       "left",
					Operator:    "or",
					Left:        4,
					Right:       5,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          3,
					AttributeID: 102,
					Value:       "right",
					Operator:    "and",
					Left:        6,
					Right:       7,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          4,
					AttributeID: 103,
					Value:       "left-left",
					Operator:    "eq",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          5,
					AttributeID: 104,
					Value:       "left-right",
					Operator:    "neq",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          6,
					AttributeID: 105,
					Value:       "right-left",
					Operator:    "gt",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
				{
					ID:          7,
					AttributeID: 106,
					Value:       "right-right",
					Operator:    "lt",
					Left:        0,
					Right:       0,
					Ctime:       now,
					Utime:       now,
				},
			},
			expected: []*domain.PolicyRule{
				{
					ID: 1,
					AttributeDefinition: domain.AttributeDefinition{
						ID: 100,
					},
					Value:    "root",
					Operator: "and",
					Ctime:    now,
					Utime:    now,
					LeftRule: &domain.PolicyRule{
						ID: 2,
						AttributeDefinition: domain.AttributeDefinition{
							ID: 101,
						},
						Value:    "left",
						Operator: "or",
						Ctime:    now,
						Utime:    now,
						LeftRule: &domain.PolicyRule{
							ID: 4,
							AttributeDefinition: domain.AttributeDefinition{
								ID: 103,
							},
							Value:    "left-left",
							Operator: "eq",
							Ctime:    now,
							Utime:    now,
						},
						RightRule: &domain.PolicyRule{
							ID: 5,
							AttributeDefinition: domain.AttributeDefinition{
								ID: 104,
							},
							Value:    "left-right",
							Operator: "neq",
							Ctime:    now,
							Utime:    now,
						},
					},
					RightRule: &domain.PolicyRule{
						ID: 3,
						AttributeDefinition: domain.AttributeDefinition{
							ID: 102,
						},
						Value:    "right",
						Operator: "and",
						Ctime:    now,
						Utime:    now,
						LeftRule: &domain.PolicyRule{
							ID: 6,
							AttributeDefinition: domain.AttributeDefinition{
								ID: 105,
							},
							Value:    "right-left",
							Operator: "gt",
							Ctime:    now,
							Utime:    now,
						},
						RightRule: &domain.PolicyRule{
							ID: 7,
							AttributeDefinition: domain.AttributeDefinition{
								ID: 106,
							},
							Value:    "right-right",
							Operator: "lt",
							Ctime:    now,
							Utime:    now,
						},
					},
				},
			},
		},
	}

	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := genDomainPolicyRules(tt.rules)
			assert.Equal(t, tt.expected, result)
		})
	}
}
