package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// BusinessConfigRepository 业务配置仓储接口
type BusinessConfigRepository interface {
	Create(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)

	Find(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, error)
	FindByID(ctx context.Context, id int64) (domain.BusinessConfig, error)

	UpdateToken(ctx context.Context, id int64, token string) error
	Update(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)

	Delete(ctx context.Context, id int64) error
}

// businessConfigRepository 业务配置仓储实现
type businessConfigRepository struct {
	businessConfigDAO dao.BusinessConfigDAO
}

// NewBusinessConfigRepository 创建业务配置仓储实例
func NewBusinessConfigRepository(businessConfigDAO dao.BusinessConfigDAO) BusinessConfigRepository {
	return &businessConfigRepository{
		businessConfigDAO: businessConfigDAO,
	}
}

func (r *businessConfigRepository) Create(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	created, err := r.businessConfigDAO.Create(ctx, r.toEntity(config))
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return r.toDomain(created), nil
}

func (r *businessConfigRepository) Find(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, error) {
	list, err := r.businessConfigDAO.Find(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(list, func(_ int, src dao.BusinessConfig) domain.BusinessConfig {
		return r.toDomain(src)
	}), nil
}

func (r *businessConfigRepository) FindByID(ctx context.Context, id int64) (domain.BusinessConfig, error) {
	config, err := r.businessConfigDAO.GetByID(ctx, id)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return r.toDomain(config), nil
}

func (r *businessConfigRepository) Update(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	err := r.businessConfigDAO.Update(ctx, r.toEntity(config))
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return config, nil
}

func (r *businessConfigRepository) UpdateToken(ctx context.Context, id int64, token string) error {
	return r.businessConfigDAO.UpdateToken(ctx, id, token)
}

func (r *businessConfigRepository) Delete(ctx context.Context, id int64) error {
	return r.businessConfigDAO.Delete(ctx, id)
}

func (r *businessConfigRepository) toEntity(bc domain.BusinessConfig) dao.BusinessConfig {
	return dao.BusinessConfig{
		ID:        bc.ID,
		OwnerID:   bc.OwnerID,
		OwnerType: bc.OwnerType,
		Name:      bc.Name,
		RateLimit: bc.RateLimit,
		Token:     bc.Token,
		Ctime:     bc.Ctime,
		Utime:     bc.Utime,
	}
}

func (r *businessConfigRepository) toDomain(bc dao.BusinessConfig) domain.BusinessConfig {
	return domain.BusinessConfig{
		ID:        bc.ID,
		OwnerID:   bc.OwnerID,
		OwnerType: bc.OwnerType,
		Name:      bc.Name,
		RateLimit: bc.RateLimit,
		Token:     bc.Token,
		Ctime:     bc.Ctime,
		Utime:     bc.Utime,
	}
}
