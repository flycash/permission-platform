package abac

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
)

type AttributeDefinitionSvc interface {
	// 返回属性定义id
	Save(ctx context.Context, bizID int64, definition domain.AttributeDefinition) (int64, error)
	First(ctx context.Context, bizID int64, id int64) (domain.AttributeDefinition, error)
	Del(ctx context.Context, bizID int64, id int64) error
	// 返回一个bizID所有的属性定义
	Find(ctx context.Context, bizID int64) (domain.BizAttrDefinition, error)
}

type attributeDefinitionSvc struct {
	repo repository.AttributeDefinitionRepository
}

func NewAttributeDefinitionSvc(repo repository.AttributeDefinitionRepository) AttributeDefinitionSvc {
	return &attributeDefinitionSvc{
		repo: repo,
	}
}

func (a *attributeDefinitionSvc) Save(ctx context.Context, bizID int64, definition domain.AttributeDefinition) (int64, error) {
	return a.repo.Save(ctx, bizID, definition)
}

func (a *attributeDefinitionSvc) First(ctx context.Context, bizID, id int64) (domain.AttributeDefinition, error) {
	return a.repo.First(ctx, bizID, id)
}

func (a *attributeDefinitionSvc) Del(ctx context.Context, bizID, id int64) error {
	return a.repo.Del(ctx, bizID, id)
}

func (a *attributeDefinitionSvc) Find(ctx context.Context, bizID int64) (domain.BizAttrDefinition, error) {
	return a.repo.Find(ctx, bizID)
}
