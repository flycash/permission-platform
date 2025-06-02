package abac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
)

type AttributeValueSvc interface {
	SaveSubjectValue(ctx context.Context, bizID, subjectID int64, val domain.AttributeValue) (int64, error)
	DeleteSubjectValue(ctx context.Context, bizID, id int64) error
	FindSubjectValueWithDefinition(ctx context.Context, bizID, subjectID int64) (domain.ABACObject, error)

	SaveResourceValue(ctx context.Context, bizID, resourceID int64, val domain.AttributeValue) (int64, error)
	DeleteResourceValue(ctx context.Context, bizID, id int64) error
	FindResourceValueWithDefinition(ctx context.Context, bizID, resourceID int64) (domain.ABACObject, error)

	SaveEnvironmentValue(ctx context.Context, bizID int64, val domain.AttributeValue) (int64, error)
	DeleteEnvironmentValue(ctx context.Context, bizID, id int64) error
	FindEnvironmentValueWithDefinition(ctx context.Context, bizID int64) (domain.ABACObject, error)
}

type attributeValueSvc struct {
	repository.AttributeValueRepository
}

func NewAttributeValueSvc(repository repository.AttributeValueRepository) AttributeValueSvc {
	return &attributeValueSvc{repository}
}
