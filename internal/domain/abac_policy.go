package domain

type Policy struct {
	ID          int64
	BizID       int64
	Name        string
	Description string
	Status      PolicyStatus
	Effect      EffectType
	Rules       []*PolicyRule
	Ctime       int64
	Utime       int64
}

type EffectType string

const (
	EffectPermit EffectType = "permit"
	EffectDeny   EffectType = "deny"
)

type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusInactive PolicyStatus = "inactive"
)

type PolicyRule struct {
	ID                  int64
	AttributeDefinition AttributeDefinition
	Value               string
	LeftRule            *PolicyRule
	RightRule           *PolicyRule
	Operator            RuleOperator
	Ctime               int64
	Utime               int64
}

type RuleOperator string

func (r RuleOperator) String() string {
	return string(r)
}

const (
	Equals         RuleOperator = "="
	NotEquals      RuleOperator = "!="
	Greater        RuleOperator = ">"
	Less           RuleOperator = "<"
	GreaterOrEqual RuleOperator = ">="
	LessOrEqual    RuleOperator = "<="
	AND            RuleOperator = "AND"
	OR             RuleOperator = "OR"
	IN             RuleOperator = "IN"
	NotIn          RuleOperator = "NOT IN"
	NOT            RuleOperator = "NOT"
)
