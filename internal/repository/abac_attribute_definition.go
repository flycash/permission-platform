package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
)

type AttributeDefinitionRepository interface {
	// 返回属性定义id
	Save(ctx context.Context, bizID int64, definition domain.AttributeDefinition) (int64, error)
	First(ctx context.Context, bizID int64, id int64) (domain.AttributeDefinition, error)
	Del(ctx context.Context, bizID int64, id int64) error
	// 返回一个bizID所有的属性定义
	Find(ctx context.Context, bizID int64) (domain.BizDefinition, error)
}

type attributeDefinitionRepository struct {
	definitionDao dao.AttributeDefinitionDAO
}

func NewAttributeDefinitionRepository(definitionDao dao.AttributeDefinitionDAO) AttributeDefinitionRepository {
	return &attributeDefinitionRepository{
		definitionDao: definitionDao,
	}
}

// 将domain.AttributeDefinition转换为dao.Definition
func toDaoDefinition(bizID int64, def domain.AttributeDefinition) dao.AttributeDefinition {
	return dao.AttributeDefinition{
		ID:             def.ID,
		BizID:          bizID,
		Name:           def.Name,
		Description:    def.Description,
		DataType:       def.DataType.String(),
		EntityType:     def.EntityType.String(),
		ValidationRule: def.ValidationRule,
		Ctime:          def.Ctime,
		Utime:          def.Utime,
	}
}

// 将dao.AttributeDefinition转换为domain.Definition
func toDomainDefinition(daoDef dao.AttributeDefinition) domain.AttributeDefinition {
	return domain.AttributeDefinition{
		ID:             daoDef.ID,
		Name:           daoDef.Name,
		Description:    daoDef.Description,
		DataType:       domain.DataType(daoDef.DataType),
		EntityType:     domain.EntityType(daoDef.EntityType),
		ValidationRule: daoDef.ValidationRule,
		Ctime:          daoDef.Ctime,
		Utime:          daoDef.Utime,
	}
}

func (a *attributeDefinitionRepository) Save(ctx context.Context, bizID int64, definition domain.AttributeDefinition) (int64, error) {
	daoDef := toDaoDefinition(bizID, definition)
	return a.definitionDao.Save(ctx, daoDef)
}

func (a *attributeDefinitionRepository) First(ctx context.Context, bizID, id int64) (domain.AttributeDefinition, error) {
	daoDef, err := a.definitionDao.First(ctx, bizID, id)
	if err != nil {
		return domain.AttributeDefinition{}, err
	}
	return toDomainDefinition(daoDef), nil
}

func (a *attributeDefinitionRepository) Del(ctx context.Context, bizID, id int64) error {
	return a.definitionDao.Del(ctx, bizID, id)
}

func (a *attributeDefinitionRepository) Find(ctx context.Context, bizID int64) (domain.BizDefinition, error) {
	daos, err := a.definitionDao.Find(ctx, bizID)
	if err != nil {
		return domain.BizDefinition{}, err
	}
	bizDef := domain.BizDefinition{
		BizID: bizID,
	}
	for _, daoDef := range daos {
		def := toDomainDefinition(daoDef)
		switch daoDef.EntityType {
		case domain.SubjectType.String():
			bizDef.SubjectAttrs = append(bizDef.SubjectAttrs, def)
		case domain.ResourceType.String():
			bizDef.ResourceAttrs = append(bizDef.ResourceAttrs, def)
		case domain.EnvironmentType.String():
			bizDef.EnvironmentAttrs = append(bizDef.EnvironmentAttrs, def)
		}
	}
	return bizDef, nil
}
