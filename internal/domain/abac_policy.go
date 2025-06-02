package domain

type Policy struct {
	ID          int64
	BizID       int64
	Name        string
	Description string
	ExecuteType ExecuteType
	Status      PolicyStatus
	Permissions []UserPermission
	Rules       []PolicyRule
	Ctime       int64
	Utime       int64
}

func (p Policy) ContainsAnyPermissions(permissionIDs []int64) bool {
	for idx := range permissionIDs {
		permissionID := permissionIDs[idx]
		for jdx := range p.Permissions {
			permission := p.Permissions[jdx]
			if permission.ID == permissionID {
				return true
			}
		}
	}
	return false
}

type ExecuteType string

const LogicType ExecuteType = "logic" // 逻辑运算符执行方法

type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusInactive PolicyStatus = "inactive"
)

type PolicyRule struct {
	ID        int64
	AttrDef   AttributeDefinition
	Value     string
	LeftRule  *PolicyRule
	RightRule *PolicyRule
	Operator  RuleOperator
	Ctime     int64
	Utime     int64
}

func (p PolicyRule) SafeLeft() PolicyRule {
	if p.LeftRule == nil {
		return PolicyRule{}
	}
	return *p.LeftRule
}

func (p PolicyRule) SafeRight() PolicyRule {
	if p.RightRule == nil {
		return PolicyRule{}
	}
	return *p.RightRule
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
	AllMatch       RuleOperator = "ALL MATCH"
	AnyMatch       RuleOperator = "ANY MATCH"
)
