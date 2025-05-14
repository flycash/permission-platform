package abac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
)

type AttributeValueSvc interface {
	SaveSubjectValue(ctx context.Context, bizID, subjectID int64, val domain.SubjectAttributeValue) (int64, error)
	DeleteSubjectValue(ctx context.Context, id int64) error
	FindSubjectValue(ctx context.Context, bizID, subjectID int64) (domain.SubjectObject, error)
	FindSubjectValueWithDefinition(ctx context.Context, bizID, subjectID int64) (domain.SubjectObject, error)

	SaveResourceValue(ctx context.Context, bizID, resourceID int64, val domain.ResourceAttributeValue) (int64, error)
	DeleteResourceValue(ctx context.Context, id int64) error
	FindResourceValue(ctx context.Context, bizID, resourceID int64) (domain.ResourceObject, error)
	FindResourceValueWithDefinition(ctx context.Context, bizID, resourceID int64) (domain.ResourceObject, error)

	SaveEnvironmentValue(ctx context.Context, bizID int64, val domain.EnvironmentAttributeValue) (int64, error)
	DeleteEnvironmentValue(ctx context.Context, id int64) error
	FindEnvironmentValue(ctx context.Context, bizID int64) (domain.EnvironmentObject, error)
	FindEnvironmentValueWithDefinition(ctx context.Context, bizID int64) (domain.EnvironmentObject, error)
}

type attributeValueSvc struct {
	repository.AttributeValueRepository
}

func NewAttributeValueSvc(repository repository.AttributeValueRepository) AttributeValueSvc {
	return &attributeValueSvc{repository}
}
