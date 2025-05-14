package repository

import (
	"context"
	"fmt"
	"regexp"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type AttributeValueRepository interface {
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

type attributeValueRepository struct {
	envDao        dao.EnvironmentAttributeDAO
	resourceDao   dao.ResourceAttributeValueDAO
	subjectDao    dao.SubjectAttributeValueDAO
	definitionDao dao.AttributeDefinitionDAO
}

func NewAttributeValueRepository(envDao dao.EnvironmentAttributeDAO,
	resourceDao dao.ResourceAttributeValueDAO,
	subjectDao dao.SubjectAttributeValueDAO,
	definitionDao dao.AttributeDefinitionDAO,
) AttributeValueRepository {
	return &attributeValueRepository{
		envDao:        envDao,
		resourceDao:   resourceDao,
		subjectDao:    subjectDao,
		definitionDao: definitionDao,
	}
}

func (a *attributeValueRepository) checkVal(ctx context.Context, bizID int64, val string, definitionID int64) error {
	definition, err := a.definitionDao.First(ctx, bizID, definitionID)
	if err != nil {
		return err
	}
	return a.matchRegex(definition.ValidationRule, val)
}

func (a *attributeValueRepository) matchRegex(pattern, input string) error {
	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		return fmt.Errorf("正则表达式语法错误: %w", err)
	}
	if !matched {
		return fmt.Errorf("填写的值不符合正则规范")
	}
	return nil
}

func (a *attributeValueRepository) SaveSubjectValue(ctx context.Context, bizID, subjectID int64, val domain.SubjectAttributeValue) (int64, error) {
	err := a.checkVal(ctx, bizID, val.Value, val.Definition.ID)
	if err != nil {
		return 0, err
	}
	daoVal := dao.SubjectAttributeValue{
		ID:          val.ID,
		BizID:       bizID,
		SubjectID:   subjectID,
		AttributeID: val.Definition.ID,
		Value:       val.Value,
	}
	return a.subjectDao.Save(ctx, daoVal)
}

func (a *attributeValueRepository) DeleteSubjectValue(ctx context.Context, id int64) error {
	return a.subjectDao.Del(ctx, id)
}

func (a *attributeValueRepository) FindSubjectValue(ctx context.Context, bizID, subjectID int64) (domain.SubjectObject, error) {
	values, err := a.subjectDao.FindBySubject(ctx, bizID, subjectID)
	if err != nil {
		return domain.SubjectObject{}, err
	}
	result := domain.SubjectObject{
		ID:    subjectID,
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.SubjectAttributeValue) domain.SubjectAttributeValue {
			return a.toDomainSubjectValue(src, dao.AttributeDefinition{
				ID: src.AttributeID,
			})
		}),
	}

	return result, nil
}

func (a *attributeValueRepository) FindSubjectValueWithDefinition(ctx context.Context, bizID, subjectID int64) (domain.SubjectObject, error) {
	values, err := a.subjectDao.FindBySubject(ctx, bizID, subjectID)
	if err != nil {
		return domain.SubjectObject{}, err
	}
	definitionIds := slice.Map(values, func(_ int, src dao.SubjectAttributeValue) int64 {
		return src.AttributeID
	})
	definitionMap, err := a.definitionDao.FindByIDs(ctx, definitionIds)
	if err != nil {
		return domain.SubjectObject{}, err
	}
	result := domain.SubjectObject{
		ID:    subjectID,
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.SubjectAttributeValue) domain.SubjectAttributeValue {
			return a.toDomainSubjectValue(src, definitionMap[src.AttributeID])
		}),
	}
	return result, nil
}

func (a *attributeValueRepository) SaveResourceValue(ctx context.Context, bizID, resourceID int64, val domain.ResourceAttributeValue) (int64, error) {
	err := a.checkVal(ctx, bizID, val.Value, val.Definition.ID)
	if err != nil {
		return 0, err
	}
	daoVal := dao.ResourceAttributeValue{
		ID:          val.ID,
		BizID:       bizID,
		ResourceID:  resourceID,
		AttributeID: val.Definition.ID,
		Value:       val.Value,
	}
	return a.resourceDao.Save(ctx, daoVal)
}

func (a *attributeValueRepository) DeleteResourceValue(ctx context.Context, id int64) error {
	return a.resourceDao.Del(ctx, id)
}

func (a *attributeValueRepository) FindResourceValue(ctx context.Context, bizID, resourceID int64) (domain.ResourceObject, error) {
	values, err := a.resourceDao.FindByResource(ctx, bizID, resourceID)
	if err != nil {
		return domain.ResourceObject{}, err
	}
	result := domain.ResourceObject{
		ID:    resourceID,
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.ResourceAttributeValue) domain.ResourceAttributeValue {
			return a.toDomainResourceValue(src, dao.AttributeDefinition{ID: src.AttributeID})
		}),
	}
	return result, nil
}

func (a *attributeValueRepository) FindResourceValueWithDefinition(ctx context.Context, bizID, resourceID int64) (domain.ResourceObject, error) {
	values, err := a.resourceDao.FindByResource(ctx, bizID, resourceID)
	if err != nil {
		return domain.ResourceObject{}, err
	}
	definitionIds := slice.Map(values, func(_ int, src dao.ResourceAttributeValue) int64 {
		return src.AttributeID
	})
	definitionMap, err := a.definitionDao.FindByIDs(ctx, definitionIds)
	if err != nil {
		return domain.ResourceObject{}, err
	}
	result := domain.ResourceObject{
		ID:    resourceID,
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.ResourceAttributeValue) domain.ResourceAttributeValue {
			return a.toDomainResourceValue(src, definitionMap[src.AttributeID])
		}),
	}
	return result, nil
}

func (a *attributeValueRepository) SaveEnvironmentValue(ctx context.Context, bizID int64, val domain.EnvironmentAttributeValue) (int64, error) {
	err := a.checkVal(ctx, bizID, val.Value, val.Definition.ID)
	if err != nil {
		return 0, err
	}
	daoVal := dao.EnvironmentAttributeValue{
		ID:          val.ID,
		BizID:       bizID,
		AttributeID: val.Definition.ID,
		Value:       val.Value,
	}
	return a.envDao.Save(ctx, daoVal)
}

func (a *attributeValueRepository) DeleteEnvironmentValue(ctx context.Context, id int64) error {
	return a.envDao.Del(ctx, id)
}

func (a *attributeValueRepository) FindEnvironmentValue(ctx context.Context, bizID int64) (domain.EnvironmentObject, error) {
	values, err := a.envDao.FindByBiz(ctx, bizID)
	if err != nil {
		return domain.EnvironmentObject{}, err
	}
	return domain.EnvironmentObject{
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.EnvironmentAttributeValue) domain.EnvironmentAttributeValue {
			return a.toDomainEnvironmentAttributeValue(src, dao.AttributeDefinition{
				ID: src.AttributeID,
			})
		}),
	}, nil
}

func (a *attributeValueRepository) FindEnvironmentValueWithDefinition(ctx context.Context, bizID int64) (domain.EnvironmentObject, error) {
	values, err := a.envDao.FindByBiz(ctx, bizID)
	if err != nil {
		return domain.EnvironmentObject{}, err
	}
	definitionIds := slice.Map(values, func(_ int, src dao.EnvironmentAttributeValue) int64 {
		return src.AttributeID
	})
	definitionMap, err := a.definitionDao.FindByIDs(ctx, definitionIds)
	if err != nil {
		return domain.EnvironmentObject{}, err
	}
	return domain.EnvironmentObject{
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.EnvironmentAttributeValue) domain.EnvironmentAttributeValue {
			return a.toDomainEnvironmentAttributeValue(src, definitionMap[src.AttributeID])
		}),
	}, nil
}

func (a *attributeValueRepository) toDomainSubjectValue(subjectVal dao.SubjectAttributeValue, definition dao.AttributeDefinition) domain.SubjectAttributeValue {
	return domain.SubjectAttributeValue{
		ID:         subjectVal.ID,
		Definition: toDomainDefinition(definition),
		Value:      subjectVal.Value,
		Ctime:      subjectVal.Ctime,
		Utime:      subjectVal.Utime,
	}
}

func (a *attributeValueRepository) toDomainResourceValue(resourceVal dao.ResourceAttributeValue, definition dao.AttributeDefinition) domain.ResourceAttributeValue {
	return domain.ResourceAttributeValue{
		ID:         resourceVal.ID,
		Definition: toDomainDefinition(definition),
		Value:      resourceVal.Value,
		Ctime:      resourceVal.Ctime,
		Utime:      resourceVal.Utime,
	}
}

func (a *attributeValueRepository) toDomainEnvironmentAttributeValue(envVal dao.EnvironmentAttributeValue, definition dao.AttributeDefinition) domain.EnvironmentAttributeValue {
	return domain.EnvironmentAttributeValue{
		ID:         envVal.ID,
		Definition: toDomainDefinition(definition),
		Value:      envVal.Value,
		Ctime:      envVal.Ctime,
		Utime:      envVal.Utime,
	}
}
