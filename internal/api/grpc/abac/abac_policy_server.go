package abac

import (
	"context"

	permissionpb "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/domain"
	abacSvc "gitee.com/flycash/permission-platform/internal/service/abac"
)

type ABACPolicyServer struct {
	permissionpb.UnsafePolicyServiceServer
	baseServer
	svc abacSvc.PolicySvc
}

func NewABACPolicyServer(svc abacSvc.PolicySvc) *ABACPolicyServer {
	return &ABACPolicyServer{
		svc: svc,
	}
}

func (s *ABACPolicyServer) Save(ctx context.Context, req *permissionpb.PolicyServiceSaveRequest) (*permissionpb.PolicyServiceSaveResponse, error) {
	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	policy := convertToDomainPolicy(req.Policy)
	policy.BizID = bizID
	id, err := s.svc.Save(ctx, policy)
	if err != nil {
		return nil, err
	}
	return &permissionpb.PolicyServiceSaveResponse{
		Id: id,
	}, nil
}

func (s *ABACPolicyServer) Delete(ctx context.Context, req *permissionpb.PolicyServiceDeleteRequest) (*permissionpb.PolicyServiceDeleteResponse, error) {
	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	err = s.svc.Delete(ctx, bizID, req.Id)
	if err != nil {
		return nil, err
	}
	return &permissionpb.PolicyServiceDeleteResponse{}, nil
}

func (s *ABACPolicyServer) First(ctx context.Context, req *permissionpb.PolicyServiceFirstRequest) (*permissionpb.PolicyServiceFirstResponse, error) {
	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	policy, err := s.svc.First(ctx, bizID, req.Id)
	if err != nil {
		return nil, err
	}
	return &permissionpb.PolicyServiceFirstResponse{
		Policy: convertToProtoPolicy(policy),
	}, nil
}

func (s *ABACPolicyServer) SaveRule(ctx context.Context, req *permissionpb.PolicyServiceSaveRuleRequest) (*permissionpb.PolicyServiceSaveRuleResponse, error) {
	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rule := convertToDomainPolicyRule(req.Rule)
	id, err := s.svc.SaveRule(ctx, bizID, req.PolicyId, rule) // Dereference the pointer
	if err != nil {
		return nil, err
	}
	return &permissionpb.PolicyServiceSaveRuleResponse{
		Id: id,
	}, nil
}

func (s *ABACPolicyServer) DeleteRule(ctx context.Context, req *permissionpb.PolicyServiceDeleteRuleRequest) (*permissionpb.PolicyServiceDeleteRuleResponse, error) {
	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	err = s.svc.DeleteRule(ctx, bizID, req.RuleId)
	if err != nil {
		return nil, err
	}
	return &permissionpb.PolicyServiceDeleteRuleResponse{}, nil
}

func (s *ABACPolicyServer) SavePermissionPolicy(ctx context.Context, req *permissionpb.PolicyServiceSavePermissionPolicyRequest) (*permissionpb.PolicyServiceSavePermissionPolicyResponse, error) {
	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	err = s.svc.SavePermissionPolicy(ctx, bizID, req.PolicyId, req.PermissionId, convertToDomainEffect(req.Effect))
	if err != nil {
		return nil, err
	}
	return &permissionpb.PolicyServiceSavePermissionPolicyResponse{}, nil
}

func (s *ABACPolicyServer) FindPolicies(ctx context.Context, req *permissionpb.PolicyServiceFindPoliciesRequest) (*permissionpb.PolicyServiceFindPoliciesResponse, error) {
	bizID, err := s.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	total, policies, err := s.svc.FindPolicies(ctx, bizID, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, err
	}
	return &permissionpb.PolicyServiceFindPoliciesResponse{
		Total:    total,
		Policies: convertToProtoPolicies(policies),
	}, nil
}

// Helper functions to convert between domain and proto types
func convertToDomainPolicy(p *permissionpb.Policy) domain.Policy {
	if p == nil {
		return domain.Policy{}
	}
	return domain.Policy{
		ID:          p.Id,
		Name:        p.Name,
		Description: p.Description,
		Status:      convertToDomainPolicyStatus(p.Status),
		Permissions: []domain.UserPermission{
			{
				Effect: domain.Effect(p.Effect),
			},
		},

		Rules: convertToDomainPolicyRules(p.Rules),
		Ctime: p.Ctime,
		Utime: p.Utime,
	}
}

func convertToProtoPolicy(p domain.Policy) *permissionpb.Policy {
	var effect domain.Effect
	if len(p.Permissions) > 0 {
		effect = p.Permissions[0].Effect
	}
	return &permissionpb.Policy{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Status:      convertToProtoPolicyStatus(p.Status),
		Effect:      convertToProtoEffect(effect),
		Rules:       convertToProtoPolicyRules(p.Rules),
		Ctime:       p.Ctime,
		Utime:       p.Utime,
	}
}

func convertToDomainPolicyRules(rules []*permissionpb.PolicyRule) []domain.PolicyRule {
	if rules == nil {
		return nil
	}
	result := make([]domain.PolicyRule, 0, len(rules))
	for _, rule := range rules {
		result = append(result, convertToDomainPolicyRule(rule))
	}
	return result
}

func convertToProtoPolicyRules(rules []domain.PolicyRule) []*permissionpb.PolicyRule {
	if rules == nil {
		return nil
	}
	result := make([]*permissionpb.PolicyRule, 0, len(rules))
	for _, rule := range rules {
		result = append(result, convertToProtoPolicyRule(rule))
	}
	return result
}

func convertToDomainPolicyRule(r *permissionpb.PolicyRule) domain.PolicyRule {
	if r == nil {
		return domain.PolicyRule{}
	}
	left := convertToDomainPolicyRule(r.LeftRule)
	right := convertToDomainPolicyRule(r.RightRule)
	return domain.PolicyRule{
		ID:        r.Id,
		AttrDef:   convertToDomainAttributeDefinition(r.AttributeDefinition),
		Value:     r.Value,
		Operator:  convertToDomainOperator(r.Operator),
		LeftRule:  &left,
		RightRule: &right,
	}
}

func convertToProtoPolicyRule(r domain.PolicyRule) *permissionpb.PolicyRule {
	return &permissionpb.PolicyRule{
		Id:                  r.ID,
		AttributeDefinition: convertToProtoAttributeDefinition(r.AttrDef),
		Value:               r.Value,
		Operator:            convertToProtoOperator(r.Operator),
		LeftRule:            convertToProtoPolicyRule(r.SafeLeft()),
		RightRule:           convertToProtoPolicyRule(r.SafeRight()),
	}
}

func convertToDomainAttributeDefinition(d *permissionpb.AttributeDefinition) domain.AttributeDefinition {
	if d == nil {
		return domain.AttributeDefinition{}
	}
	return domain.AttributeDefinition{
		ID:             d.Id,
		Name:           d.Name,
		Description:    d.Description,
		DataType:       convertToDomainDataType(d.DataType),
		EntityType:     convertToDomainEntityType(d.EntityType),
		ValidationRule: d.ValidationRule,
		Ctime:          d.Ctime,
		Utime:          d.Utime,
	}
}

func convertToProtoAttributeDefinition(d domain.AttributeDefinition) *permissionpb.AttributeDefinition {
	return &permissionpb.AttributeDefinition{
		Id:             d.ID,
		Name:           d.Name,
		Description:    d.Description,
		DataType:       convertToProtoDataType(d.DataType),
		EntityType:     convertToProtoEntityType(d.EntityType),
		ValidationRule: d.ValidationRule,
		Ctime:          d.Ctime,
		Utime:          d.Utime,
	}
}

func convertToDomainPolicyStatus(s permissionpb.PolicyStatus) domain.PolicyStatus {
	switch s {
	case permissionpb.PolicyStatus_POLICY_STATUS_ACTIVE:
		return domain.PolicyStatusActive
	case permissionpb.PolicyStatus_POLICY_STATUS_INACTIVE:
		return domain.PolicyStatusInactive
	default:
		return ""
	}
}

func convertToProtoPolicyStatus(s domain.PolicyStatus) permissionpb.PolicyStatus {
	switch s {
	case domain.PolicyStatusActive:
		return permissionpb.PolicyStatus_POLICY_STATUS_ACTIVE
	case domain.PolicyStatusInactive:
		return permissionpb.PolicyStatus_POLICY_STATUS_INACTIVE
	default:
		return permissionpb.PolicyStatus_POLICY_STATUS_UNKNOWN
	}
}

func convertToDomainEffect(e permissionpb.Effect) domain.Effect {
	switch e {
	case permissionpb.Effect_EFFECT_ALLOW:
		return domain.EffectAllow
	case permissionpb.Effect_EFFECT_DENY:
		return domain.EffectDeny
	default:
		return "" // Use zero value instead of undefined constant
	}
}

func convertToProtoEffect(e domain.Effect) permissionpb.Effect {
	switch e {
	case domain.EffectAllow:
		return permissionpb.Effect_EFFECT_ALLOW
	case domain.EffectDeny:
		return permissionpb.Effect_EFFECT_DENY
	default:
		return permissionpb.Effect_EFFECT_UNKNOWN
	}
}

func convertToDomainDataType(d permissionpb.DataType) domain.DataType {
	switch d {
	case permissionpb.DataType_DATA_TYPE_STRING:
		return domain.DataTypeString
	case permissionpb.DataType_DATA_TYPE_NUMBER:
		return domain.DataTypeNumber
	case permissionpb.DataType_DATA_TYPE_BOOLEAN:
		return domain.DataTypeBoolean
	case permissionpb.DataType_DATA_TYPE_FLOAT:
		return domain.DataTypeFloat
	case permissionpb.DataType_DATA_TYPE_DATETIME:
		return domain.DataTypeDatetime
	default:
		return domain.DataType("")
	}
}

func convertToProtoDataType(d domain.DataType) permissionpb.DataType {
	switch d {
	case domain.DataTypeString:
		return permissionpb.DataType_DATA_TYPE_STRING
	case domain.DataTypeNumber:
		return permissionpb.DataType_DATA_TYPE_NUMBER
	case domain.DataTypeBoolean:
		return permissionpb.DataType_DATA_TYPE_BOOLEAN
	case domain.DataTypeFloat:
		return permissionpb.DataType_DATA_TYPE_FLOAT
	case domain.DataTypeDatetime:
		return permissionpb.DataType_DATA_TYPE_DATETIME
	default:
		return permissionpb.DataType_DATA_TYPE_UNKNOWN
	}
}

func convertToDomainEntityType(e permissionpb.EntityType) domain.EntityType {
	switch e {
	case permissionpb.EntityType_ENTITY_TYPE_SUBJECT:
		return domain.EntityTypeSubject
	case permissionpb.EntityType_ENTITY_TYPE_RESOURCE:
		return domain.EntityTypeResource
	case permissionpb.EntityType_ENTITY_TYPE_ENVIRONMENT:
		return domain.EntityTypeEnvironment
	default:
		return ""
	}
}

func convertToProtoEntityType(e domain.EntityType) permissionpb.EntityType {
	switch e {
	case domain.EntityTypeSubject:
		return permissionpb.EntityType_ENTITY_TYPE_SUBJECT
	case domain.EntityTypeResource:
		return permissionpb.EntityType_ENTITY_TYPE_RESOURCE
	case domain.EntityTypeEnvironment:
		return permissionpb.EntityType_ENTITY_TYPE_ENVIRONMENT
	default:
		return permissionpb.EntityType_ENTITY_TYPE_UNKNOWN
	}
}

func convertToDomainOperator(o permissionpb.RuleOperator) domain.RuleOperator {
	switch o {
	case permissionpb.RuleOperator_RULE_OPERATOR_EQUALS:
		return domain.Equals
	case permissionpb.RuleOperator_RULE_OPERATOR_NOT_EQUALS:
		return domain.NotEquals
	case permissionpb.RuleOperator_RULE_OPERATOR_GREATER:
		return domain.Greater
	case permissionpb.RuleOperator_RULE_OPERATOR_LESS:
		return domain.Less
	case permissionpb.RuleOperator_RULE_OPERATOR_GREATER_OR_EQUAL:
		return domain.GreaterOrEqual
	case permissionpb.RuleOperator_RULE_OPERATOR_LESS_OR_EQUAL:
		return domain.LessOrEqual
	case permissionpb.RuleOperator_RULE_OPERATOR_AND:
		return domain.AND
	case permissionpb.RuleOperator_RULE_OPERATOR_OR:
		return domain.OR
	case permissionpb.RuleOperator_RULE_OPERATOR_IN:
		return domain.IN
	case permissionpb.RuleOperator_RULE_OPERATOR_NOT_IN:
		return domain.NotIn
	case permissionpb.RuleOperator_RULE_OPERATOR_NOT:
		return domain.NOT
	default:
		return domain.RuleOperator("")
	}
}

func convertToProtoOperator(o domain.RuleOperator) permissionpb.RuleOperator {
	switch o {
	case domain.Equals:
		return permissionpb.RuleOperator_RULE_OPERATOR_EQUALS
	case domain.NotEquals:
		return permissionpb.RuleOperator_RULE_OPERATOR_NOT_EQUALS
	case domain.Greater:
		return permissionpb.RuleOperator_RULE_OPERATOR_GREATER
	case domain.Less:
		return permissionpb.RuleOperator_RULE_OPERATOR_LESS
	case domain.GreaterOrEqual:
		return permissionpb.RuleOperator_RULE_OPERATOR_GREATER_OR_EQUAL
	case domain.LessOrEqual:
		return permissionpb.RuleOperator_RULE_OPERATOR_LESS_OR_EQUAL
	case domain.AND:
		return permissionpb.RuleOperator_RULE_OPERATOR_AND
	case domain.OR:
		return permissionpb.RuleOperator_RULE_OPERATOR_OR
	case domain.IN:
		return permissionpb.RuleOperator_RULE_OPERATOR_IN
	case domain.NotIn:
		return permissionpb.RuleOperator_RULE_OPERATOR_NOT_IN
	case domain.NOT:
		return permissionpb.RuleOperator_RULE_OPERATOR_NOT
	default:
		return permissionpb.RuleOperator_RULE_OPERATOR_UNKNOWN
	}
}

func convertToProtoPolicies(policies []domain.Policy) []*permissionpb.Policy {
	if policies == nil {
		return nil
	}
	result := make([]*permissionpb.Policy, 0, len(policies))
	for _, policy := range policies {
		result = append(result, convertToProtoPolicy(policy))
	}
	return result
}
