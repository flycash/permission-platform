package abac

import (
	"context"

	permissionpb "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/domain"
	abacSvc "gitee.com/flycash/permission-platform/internal/service/abac"
)

type ABACAttributeValServer struct {
	permissionpb.UnsafeAttributeValueServiceServer
	baseServer
	svc abacSvc.AttributeValueSvc
}

func NewABACAttributeValServer(svc abacSvc.AttributeValueSvc) *ABACAttributeValServer {
	return &ABACAttributeValServer{svc: svc}
}

func (a *ABACAttributeValServer) SaveSubjectValue(ctx context.Context, request *permissionpb.AttributeValueServiceSaveSubjectValueRequest) (*permissionpb.AttributeValueServiceSaveSubjectValueResponse, error) {
	bizID, err := a.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	val := domain.SubjectAttributeValue{
		Definition: convertToDomainAttributeDefinition(request.Value.Definition),
		Value:      request.Value.Value,
	}
	id, err := a.svc.SaveSubjectValue(ctx, bizID, request.SubjectId, val)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceSaveSubjectValueResponse{
		Id: id,
	}, nil
}

func (a *ABACAttributeValServer) DeleteSubjectValue(ctx context.Context, request *permissionpb.AttributeValueServiceDeleteSubjectValueRequest) (*permissionpb.AttributeValueServiceDeleteSubjectValueResponse, error) {
	err := a.svc.DeleteSubjectValue(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceDeleteSubjectValueResponse{}, nil
}

func (a *ABACAttributeValServer) FindSubjectValueWithDefinition(ctx context.Context, request *permissionpb.AttributeValueServiceFindSubjectValueWithDefinitionRequest) (*permissionpb.AttributeValueServiceFindSubjectValueWithDefinitionResponse, error) {
	bizID, err := a.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	subject, err := a.svc.FindSubjectValueWithDefinition(ctx, bizID, request.SubjectId)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceFindSubjectValueWithDefinitionResponse{
		Subject: &permissionpb.SubjectObject{
			Id:              subject.ID,
			AttributeValues: convertToSubjectAttributeValues(subject.AttributeValues),
		},
	}, nil
}

func (a *ABACAttributeValServer) SaveResourceValue(ctx context.Context, request *permissionpb.AttributeValueServiceSaveResourceValueRequest) (*permissionpb.AttributeValueServiceSaveResourceValueResponse, error) {
	bizID, err := a.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	val := domain.ResourceAttributeValue{
		Definition: convertToDomainAttributeDefinition(request.Value.Definition),
		Value:      request.Value.Value,
	}
	id, err := a.svc.SaveResourceValue(ctx, bizID, request.ResourceId, val)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceSaveResourceValueResponse{
		Id: id,
	}, nil
}

func (a *ABACAttributeValServer) DeleteResourceValue(ctx context.Context, request *permissionpb.AttributeValueServiceDeleteResourceValueRequest) (*permissionpb.AttributeValueServiceDeleteResourceValueResponse, error) {
	err := a.svc.DeleteResourceValue(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceDeleteResourceValueResponse{}, nil
}

func (a *ABACAttributeValServer) FindResourceValueWithDefinition(ctx context.Context, request *permissionpb.AttributeValueServiceFindResourceValueWithDefinitionRequest) (*permissionpb.AttributeValueServiceFindResourceValueWithDefinitionResponse, error) {
	bizID, err := a.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	resource, err := a.svc.FindResourceValueWithDefinition(ctx, bizID, request.ResourceId)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceFindResourceValueWithDefinitionResponse{
		Resource: &permissionpb.ResourceObject{
			Id:              resource.ID,
			AttributeValues: convertToResourceAttributeValues(resource.AttributeValues),
		},
	}, nil
}

func (a *ABACAttributeValServer) SaveEnvironmentValue(ctx context.Context, request *permissionpb.AttributeValueServiceSaveEnvironmentValueRequest) (*permissionpb.AttributeValueServiceSaveEnvironmentValueResponse, error) {
	bizID, err := a.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	val := domain.EnvironmentAttributeValue{
		Definition: convertToDomainAttributeDefinition(request.Value.Definition),
		Value:      request.Value.Value,
	}
	id, err := a.svc.SaveEnvironmentValue(ctx, bizID, val)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceSaveEnvironmentValueResponse{
		Id: id,
	}, nil
}

func (a *ABACAttributeValServer) DeleteEnvironmentValue(ctx context.Context, request *permissionpb.AttributeValueServiceDeleteEnvironmentValueRequest) (*permissionpb.AttributeValueServiceDeleteEnvironmentValueResponse, error) {
	err := a.svc.DeleteEnvironmentValue(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceDeleteEnvironmentValueResponse{}, nil
}

func (a *ABACAttributeValServer) FindEnvironmentValueWithDefinition(ctx context.Context, request *permissionpb.AttributeValueServiceFindEnvironmentValueWithDefinitionRequest) (*permissionpb.AttributeValueServiceFindEnvironmentValueWithDefinitionResponse, error) {
	bizID, err := a.getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	env, err := a.svc.FindEnvironmentValueWithDefinition(ctx, bizID)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeValueServiceFindEnvironmentValueWithDefinitionResponse{
		Environment: &permissionpb.EnvironmentObject{
			AttributeValues: convertToEnvironmentAttributeValues(env.AttributeValues),
		},
	}, nil
}

func convertToSubjectAttributeValues(values []domain.SubjectAttributeValue) []*permissionpb.SubjectAttributeValue {
	result := make([]*permissionpb.SubjectAttributeValue, 0, len(values))
	for _, v := range values {
		result = append(result, &permissionpb.SubjectAttributeValue{
			Id:         v.ID,
			Definition: convertToProtoAttributeDefinition(v.Definition),
			Value:      v.Value,
			Ctime:      v.Ctime,
			Utime:      v.Utime,
		})
	}
	return result
}

func convertToResourceAttributeValues(values []domain.ResourceAttributeValue) []*permissionpb.ResourceAttributeValue {
	result := make([]*permissionpb.ResourceAttributeValue, 0, len(values))
	for _, v := range values {
		result = append(result, &permissionpb.ResourceAttributeValue{
			Id:         v.ID,
			Definition: convertToProtoAttributeDefinition(v.Definition),
			Value:      v.Value,
			Ctime:      v.Ctime,
			Utime:      v.Utime,
		})
	}
	return result
}

func convertToEnvironmentAttributeValues(values []domain.EnvironmentAttributeValue) []*permissionpb.EnvironmentAttributeValue {
	result := make([]*permissionpb.EnvironmentAttributeValue, 0, len(values))
	for _, v := range values {
		result = append(result, &permissionpb.EnvironmentAttributeValue{
			Id:         v.ID,
			Definition: convertToProtoAttributeDefinition(v.Definition),
			Value:      v.Value,
			Ctime:      v.Ctime,
			Utime:      v.Utime,
		})
	}
	return result
}
