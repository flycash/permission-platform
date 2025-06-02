package repository

import (
	"context"
	"fmt"
	"regexp"

	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/gotomicro/ego/core/elog"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type AttributeValueRepository interface {
	SaveSubjectValue(ctx context.Context, bizID, subjectID int64, val domain.AttributeValue) (int64, error)
	DeleteSubjectValue(ctx context.Context, bizID, id int64) error
	FindSubjectValue(ctx context.Context, bizID, subjectID int64) (domain.ABACObject, error)
	FindSubjectValueWithDefinition(ctx context.Context, bizID, subjectID int64) (domain.ABACObject, error)

	SaveResourceValue(ctx context.Context, bizID, resourceID int64, val domain.AttributeValue) (int64, error)
	DeleteResourceValue(ctx context.Context, bizID, id int64) error
	FindResourceValue(ctx context.Context, bizID, resourceID int64) (domain.ABACObject, error)
	FindResourceValueWithDefinition(ctx context.Context, bizID, resourceID int64) (domain.ABACObject, error)

	SaveEnvironmentValue(ctx context.Context, bizID int64, val domain.AttributeValue) (int64, error)
	DeleteEnvironmentValue(ctx context.Context, bizID, id int64) error
	FindEnvironmentValue(ctx context.Context, bizID int64) (domain.ABACObject, error)
	FindEnvironmentValueWithDefinition(ctx context.Context, bizID int64) (domain.ABACObject, error)
}

type attributeValueRepository struct {
	envDao        dao.EnvironmentAttributeDAO
	resourceDao   dao.ResourceAttributeValueDAO
	subjectDao    dao.SubjectAttributeValueDAO
	definitionDao dao.AttributeDefinitionDAO
	redisCache    cache.ABACAttributeValCache
	localCache    cache.ABACAttributeValCache
	logger        *elog.Component
}

func NewAttributeValueRepository(envDao dao.EnvironmentAttributeDAO,
	resourceDao dao.ResourceAttributeValueDAO,
	subjectDao dao.SubjectAttributeValueDAO,
	definitionDao dao.AttributeDefinitionDAO,
	redisCache, localCache cache.ABACAttributeValCache,
) AttributeValueRepository {
	return &attributeValueRepository{
		envDao:        envDao,
		resourceDao:   resourceDao,
		subjectDao:    subjectDao,
		definitionDao: definitionDao,
		localCache:    localCache,
		redisCache:    redisCache,
		logger:        elog.DefaultLogger,
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

func (a *attributeValueRepository) SaveSubjectValue(ctx context.Context, bizID, subjectID int64, val domain.AttributeValue) (int64, error) {
	err := a.checkVal(ctx, bizID, val.Value, val.Definition.ID)
	if err != nil {
		return 0, err
	}
	daoVal := dao.SubjectAttributeValue{
		ID:        val.ID,
		BizID:     bizID,
		SubjectID: subjectID,
		AttrDefID: val.Definition.ID,
		Value:     val.Value,
	}
	id, err := a.subjectDao.Save(ctx, daoVal)
	if err != nil {
		return 0, err
	}
	a.setSubAttributeVal(ctx, bizID, subjectID)
	return id, nil
}

func (a *attributeValueRepository) DeleteSubjectValue(ctx context.Context, bizID, id int64) error {
	err := a.subjectDao.Del(ctx, id)
	if err != nil {
		return err
	}
	a.setSubAttributeVal(ctx, bizID, id)
	return nil
}

func (a *attributeValueRepository) FindSubjectValue(ctx context.Context, bizID, subjectID int64) (domain.ABACObject, error) {
	// 本地缓存
	obj, err := a.localCache.GetAttrSubObj(ctx, bizID, subjectID)
	if err == nil {
		return obj, nil
	}
	// redis缓存
	obj, err = a.redisCache.GetAttrSubObj(ctx, bizID, subjectID)
	if err == nil {
		return obj, nil
	}
	obj, err = a.getSubAttributeVal(ctx, bizID, subjectID)
	if err != nil {
		return domain.ABACObject{}, err
	}
	// 设置redis缓存
	rerr := a.redisCache.SetAttrSubObj(ctx, []domain.ABACObject{obj})
	if rerr != nil {
		a.logger.Error("环境属性保存到redis失败", elog.FieldErr(err), elog.Int64("bizID", bizID))
	}
	return obj, nil
}

func (a *attributeValueRepository) FindSubjectValueWithDefinition(ctx context.Context, bizID, subjectID int64) (domain.ABACObject, error) {
	values, err := a.subjectDao.FindBySubject(ctx, bizID, subjectID)
	if err != nil {
		return domain.ABACObject{}, err
	}
	definitionIds := slice.Map(values, func(_ int, src dao.SubjectAttributeValue) int64 {
		return src.AttrDefID
	})
	definitionMap, err := a.definitionDao.FindByIDs(ctx, definitionIds)
	if err != nil {
		return domain.ABACObject{}, err
	}
	result := domain.ABACObject{
		ID:    subjectID,
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.SubjectAttributeValue) domain.AttributeValue {
			return a.toDomainSubjectValue(src, definitionMap[src.AttrDefID])
		}),
	}
	return result, nil
}

func (a *attributeValueRepository) SaveResourceValue(ctx context.Context, bizID, resourceID int64, val domain.AttributeValue) (int64, error) {
	err := a.checkVal(ctx, bizID, val.Value, val.Definition.ID)
	if err != nil {
		return 0, err
	}
	daoVal := dao.ResourceAttributeValue{
		ID:         val.ID,
		BizID:      bizID,
		ResourceID: resourceID,
		AttrDefID:  val.Definition.ID,
		Value:      val.Value,
	}
	id, err := a.resourceDao.Save(ctx, daoVal)
	if err != nil {
		return 0, err
	}
	a.setResAttributeVal(ctx, bizID, resourceID)
	return id, nil
}

func (a *attributeValueRepository) DeleteResourceValue(ctx context.Context, bizID, id int64) error {
	err := a.resourceDao.Del(ctx, id)
	if err != nil {
		return err
	}
	a.setResAttributeVal(ctx, bizID, id)
	return nil
}

func (a *attributeValueRepository) FindResourceValue(ctx context.Context, bizID, resourceID int64) (domain.ABACObject, error) {
	// 本地缓存
	obj, err := a.localCache.GetAttrResObj(ctx, bizID, resourceID)
	if err == nil {
		return obj, nil
	}
	// redis缓存
	obj, err = a.redisCache.GetAttrResObj(ctx, bizID, resourceID)
	if err == nil {
		return obj, nil
	}
	obj, err = a.getResAttributeVal(ctx, bizID, resourceID)
	if err != nil {
		return domain.ABACObject{}, err
	}
	// 设置redis缓存
	rerr := a.redisCache.SetAttrResObj(ctx, []domain.ABACObject{obj})
	if rerr != nil {
		a.logger.Error("环境属性保存到redis失败", elog.FieldErr(err), elog.Int64("bizID", bizID))
	}
	return obj, nil
}

func (a *attributeValueRepository) FindResourceValueWithDefinition(ctx context.Context, bizID, resourceID int64) (domain.ABACObject, error) {
	values, err := a.resourceDao.FindByResource(ctx, bizID, resourceID)
	if err != nil {
		return domain.ABACObject{}, err
	}
	definitionIds := slice.Map(values, func(_ int, src dao.ResourceAttributeValue) int64 {
		return src.AttrDefID
	})
	definitionMap, err := a.definitionDao.FindByIDs(ctx, definitionIds)
	if err != nil {
		return domain.ABACObject{}, err
	}
	result := domain.ABACObject{
		ID:    resourceID,
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.ResourceAttributeValue) domain.AttributeValue {
			return a.toDomainResourceValue(src, definitionMap[src.AttrDefID])
		}),
	}
	return result, nil
}

func (a *attributeValueRepository) SaveEnvironmentValue(ctx context.Context, bizID int64, val domain.AttributeValue) (int64, error) {
	err := a.checkVal(ctx, bizID, val.Value, val.Definition.ID)
	if err != nil {
		return 0, err
	}
	daoVal := dao.EnvironmentAttributeValue{
		ID:        val.ID,
		BizID:     bizID,
		AttrDefID: val.Definition.ID,
		Value:     val.Value,
	}
	id, err := a.envDao.Save(ctx, daoVal)
	if err != nil {
		return 0, err
	}
	a.setEnvVal(ctx, bizID)
	return id, nil
}

func (a *attributeValueRepository) DeleteEnvironmentValue(ctx context.Context, bizID, id int64) error {
	err := a.envDao.Del(ctx, id)
	if err != nil {
		return err
	}
	a.setEnvVal(ctx, bizID)
	return nil
}

func (a *attributeValueRepository) FindEnvironmentValue(ctx context.Context, bizID int64) (domain.ABACObject, error) {
	obj, err := a.redisCache.GetAttrEnvObj(ctx, bizID)
	if err == nil {
		return obj, nil
	}
	obj, err = a.getEnvVal(ctx, bizID)
	if err != nil {
		return domain.ABACObject{}, err
	}
	err = a.redisCache.SetAttrEnvObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		a.logger.Error("环境属性保存到redis失败", elog.FieldErr(err), elog.Int64("bizID", bizID))
	}
	return obj, nil
}

func (a *attributeValueRepository) FindEnvironmentValueWithDefinition(ctx context.Context, bizID int64) (domain.ABACObject, error) {
	values, err := a.envDao.FindByBiz(ctx, bizID)
	if err != nil {
		return domain.ABACObject{}, err
	}
	definitionIds := slice.Map(values, func(_ int, src dao.EnvironmentAttributeValue) int64 {
		return src.AttrDefID
	})
	definitionMap, err := a.definitionDao.FindByIDs(ctx, definitionIds)
	if err != nil {
		return domain.ABACObject{}, err
	}
	return domain.ABACObject{
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.EnvironmentAttributeValue) domain.AttributeValue {
			return a.toDomainEnvironmentAttributeValue(src, definitionMap[src.AttrDefID])
		}),
	}, nil
}

func (a *attributeValueRepository) toDomainSubjectValue(subjectVal dao.SubjectAttributeValue, definition dao.AttributeDefinition) domain.AttributeValue {
	return domain.AttributeValue{
		ID:         subjectVal.ID,
		Definition: toDomainDefinition(definition),
		Value:      subjectVal.Value,
		Ctime:      subjectVal.Ctime,
		Utime:      subjectVal.Utime,
	}
}

func (a *attributeValueRepository) toDomainResourceValue(resourceVal dao.ResourceAttributeValue, definition dao.AttributeDefinition) domain.AttributeValue {
	return domain.AttributeValue{
		ID:         resourceVal.ID,
		Definition: toDomainDefinition(definition),
		Value:      resourceVal.Value,
		Ctime:      resourceVal.Ctime,
		Utime:      resourceVal.Utime,
	}
}

func (a *attributeValueRepository) toDomainEnvironmentAttributeValue(envVal dao.EnvironmentAttributeValue, definition dao.AttributeDefinition) domain.AttributeValue {
	return domain.AttributeValue{
		ID:         envVal.ID,
		Definition: toDomainDefinition(definition),
		Value:      envVal.Value,
		Ctime:      envVal.Ctime,
		Utime:      envVal.Utime,
	}
}

func (a *attributeValueRepository) getResAttributeVal(ctx context.Context, bizID int64, id int64) (domain.ABACObject, error) {
	values, err := a.resourceDao.FindByResource(ctx, bizID, id)
	if err != nil {
		return domain.ABACObject{}, err
	}
	result := domain.ABACObject{
		ID:    id,
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.ResourceAttributeValue) domain.AttributeValue {
			return a.toDomainResourceValue(src, dao.AttributeDefinition{ID: src.AttrDefID})
		}),
	}
	return result, nil
}

func (a *attributeValueRepository) getSubAttributeVal(ctx context.Context, bizID int64, id int64) (domain.ABACObject, error) {
	values, err := a.subjectDao.FindBySubject(ctx, bizID, id)
	if err != nil {
		return domain.ABACObject{}, err
	}
	result := domain.ABACObject{
		ID:    id,
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.SubjectAttributeValue) domain.AttributeValue {
			return a.toDomainSubjectValue(src, dao.AttributeDefinition{
				ID: src.AttrDefID,
			})
		}),
	}
	return result, nil
}

func (a *attributeValueRepository) getEnvVal(ctx context.Context, bizID int64,
) (domain.ABACObject, error) {
	values, err := a.envDao.FindByBiz(ctx, bizID)
	if err != nil {
		return domain.ABACObject{}, err
	}
	return domain.ABACObject{
		BizID: bizID,
		AttributeValues: slice.Map(values, func(_ int, src dao.EnvironmentAttributeValue) domain.AttributeValue {
			return a.toDomainEnvironmentAttributeValue(src, dao.AttributeDefinition{
				ID: src.AttrDefID,
			})
		}),
	}, nil
}

func (a *attributeValueRepository) setResAttributeVal(ctx context.Context, bizID int64, id int64) {
	obj, err := a.getResAttributeVal(ctx, bizID, id)
	if err != nil {
		a.logger.Error("获取资源属性失败", elog.FieldErr(err), elog.Int64("bizID", bizID), elog.Int64("resID", id))
		return
	}
	err = a.redisCache.SetAttrResObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		a.logger.Error("资源属性保存到redis失败", elog.FieldErr(err), elog.Int64("bizID", bizID), elog.Int64("resID", id))
		return
	}
	err = a.localCache.SetAttrResObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		a.logger.Error("资源属性保存到redis失败", elog.FieldErr(err), elog.Int64("bizID", bizID), elog.Int64("resID", id))
		return
	}
}

func (a *attributeValueRepository) setSubAttributeVal(ctx context.Context, bizID int64, id int64) {
	obj, err := a.getSubAttributeVal(ctx, bizID, id)
	if err != nil {
		a.logger.Error("获取主体属性失败", elog.FieldErr(err), elog.Int64("bizID", bizID), elog.Int64("resID", id))
		return
	}
	err = a.redisCache.SetAttrSubObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		a.logger.Error("主体属性保存到redis失败", elog.FieldErr(err), elog.Int64("bizID", bizID), elog.Int64("subjectID", id))
		return
	}
	err = a.localCache.SetAttrSubObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		a.logger.Error("主体属性保存到redis失败", elog.FieldErr(err), elog.Int64("bizID", bizID), elog.Int64("subjectID", id))
		return
	}
}

func (a *attributeValueRepository) setEnvVal(ctx context.Context, bizID int64) {
	obj, err := a.getEnvVal(ctx, bizID)
	if err != nil {
		a.logger.Error("获取环境属性失败", elog.FieldErr(err), elog.Int64("bizID", bizID))
		return
	}
	err = a.redisCache.SetAttrEnvObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		a.logger.Error("环境属性保存到redis失败", elog.FieldErr(err), elog.Int64("bizID", bizID))
		return
	}
	err = a.localCache.SetAttrEnvObj(ctx, []domain.ABACObject{obj})
	if err != nil {
		a.logger.Error("环境属性保存到local失败", elog.FieldErr(err), elog.Int64("bizID", bizID))
		return
	}
}
