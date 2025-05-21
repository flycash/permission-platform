package grpc

import (
	"context"

	permissionpb "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/internal/domain"
	abacSvc "gitee.com/flycash/permission-platform/internal/service/abac"
)

type ABACAttributeDefinitionServer struct {
	permissionpb.UnsafeAttributeDefinitionServiceServer
	svc abacSvc.AttributeDefinitionSvc
}

func NewABACAttributeDefinitionServer(svc abacSvc.AttributeDefinitionSvc) *ABACAttributeDefinitionServer {
	return &ABACAttributeDefinitionServer{svc: svc}
}

func (a *ABACAttributeDefinitionServer) Save(ctx context.Context, request *permissionpb.AttributeDefinitionServiceSaveRequest) (*permissionpb.AttributeDefinitionServiceSaveResponse, error) {
	bizID, err := getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	def := convertToDomainAttributeDefinition(request.Definition)
	id, err := a.svc.Save(ctx, bizID, def)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeDefinitionServiceSaveResponse{Id: id}, nil
}

func (a *ABACAttributeDefinitionServer) First(ctx context.Context, request *permissionpb.AttributeDefinitionServiceFirstRequest) (*permissionpb.AttributeDefinitionServiceFirstResponse, error) {
	bizID, err := getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	def, err := a.svc.First(ctx, bizID, request.Id)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeDefinitionServiceFirstResponse{
		Definition: convertToProtoAttributeDefinition(def),
	}, nil
}

func (a *ABACAttributeDefinitionServer) Delete(ctx context.Context, request *permissionpb.AttributeDefinitionServiceDeleteRequest) (*permissionpb.AttributeDefinitionServiceDeleteResponse, error) {
	bizID, err := getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	err = a.svc.Del(ctx, bizID, request.Id)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeDefinitionServiceDeleteResponse{}, nil
}

func (a *ABACAttributeDefinitionServer) Find(ctx context.Context, request *permissionpb.AttributeDefinitionServiceFindRequest) (*permissionpb.AttributeDefinitionServiceFindResponse, error) {
	bizID, err := getBizIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bizDef, err := a.svc.Find(ctx, bizID)
	if err != nil {
		return nil, err
	}
	return &permissionpb.AttributeDefinitionServiceFindResponse{
		BizDefinition: convertToProtoBizDefinition(bizDef),
	}, nil
}

func convertToProtoBizDefinition(d domain.BizAttrDefinition) *permissionpb.BizDefinition {
	return &permissionpb.BizDefinition{
		SubjectAttrs:     convertToProtoAttributeDefinitions(d.SubjectAttrDefs),
		ResourceAttrs:    convertToProtoAttributeDefinitions(d.ResourceAttrDefs),
		EnvironmentAttrs: convertToProtoAttributeDefinitions(d.EnvironmentAttrDefs),
	}
}

func convertToProtoAttributeDefinitions(defs []domain.AttributeDefinition) []*permissionpb.AttributeDefinition {
	res := make([]*permissionpb.AttributeDefinition, 0, len(defs))
	for _, def := range defs {
		res = append(res, convertToProtoAttributeDefinition(def))
	}
	return res
}
