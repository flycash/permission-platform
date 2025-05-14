package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// ResourceRepository 资源仓储接口
type ResourceRepository interface {
	Create(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Resource, error)
	UpdateByBizIDAndID(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindByBizIDAndTypeAndKey(ctx context.Context, bizID int64, resourceType, resourceKey string, offset, limit int) ([]domain.Resource, error)
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, error)
}

// resourceRepository 资源仓储实现
type resourceRepository struct {
	resourceDAO dao.ResourceDAO
}

// NewResourceRepository 创建资源仓储实例
func NewResourceRepository(resourceDAO dao.ResourceDAO) ResourceRepository {
	return &resourceRepository{
		resourceDAO: resourceDAO,
	}
}

func (r *resourceRepository) Create(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	created, err := r.resourceDAO.Create(ctx, r.toEntity(resource))
	if err != nil {
		return domain.Resource{}, err
	}
	return r.toDomain(created), nil
}

func (r *resourceRepository) FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Resource, error) {
	resource, err := r.resourceDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.Resource{}, err
	}
	return r.toDomain(resource), nil
}

func (r *resourceRepository) UpdateByBizIDAndID(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	err := r.resourceDAO.UpdateByBizIDAndID(ctx, r.toEntity(resource))
	if err != nil {
		return domain.Resource{}, err
	}
	return resource, nil
}

func (r *resourceRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.resourceDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *resourceRepository) FindByBizIDAndTypeAndKey(ctx context.Context, bizID int64, resourceType, resourceKey string, offset, limit int) ([]domain.Resource, error) {
	resources, err := r.resourceDAO.FindByBizIDAndTypeAndKey(ctx, bizID, resourceType, resourceKey, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(resources, func(_ int, src dao.Resource) domain.Resource {
		return r.toDomain(src)
	}), nil
}

func (r *resourceRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Resource, error) {
	resources, err := r.resourceDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(resources, func(_ int, src dao.Resource) domain.Resource {
		return r.toDomain(src)
	}), nil
}

func (r *resourceRepository) toEntity(res domain.Resource) dao.Resource {
	return dao.Resource{
		ID:          res.ID,
		BizID:       res.BizID,
		Type:        res.Type,
		Key:         res.Key,
		Name:        res.Name,
		Description: res.Description,
		Metadata:    res.Metadata,
		Ctime:       res.Ctime,
		Utime:       res.Utime,
	}
}

func (r *resourceRepository) toDomain(res dao.Resource) domain.Resource {
	return domain.Resource{
		ID:          res.ID,
		BizID:       res.BizID,
		Type:        res.Type,
		Key:         res.Key,
		Name:        res.Name,
		Description: res.Description,
		Metadata:    res.Metadata,
		Ctime:       res.Ctime,
		Utime:       res.Utime,
	}
}
