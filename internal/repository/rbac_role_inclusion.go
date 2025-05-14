package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// RoleInclusionRepository 角色包含关系仓储接口
type RoleInclusionRepository interface {
	Create(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error)

	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, error)
	FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error)
	FindByBizIDAndIncludingRoleIDs(ctx context.Context, bizID int64, includingRoleIDs []int64, offset, limit int) ([]domain.RoleInclusion, error)
	FindByBizIDAndIncludedRoleIDs(ctx context.Context, bizID int64, includedRoleIDs []int64, offset, limit int) ([]domain.RoleInclusion, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
}

// roleInclusionRepository 角色包含关系仓储实现
type roleInclusionRepository struct {
	roleInclusionDAO dao.RoleInclusionDAO
}

// NewRoleInclusionRepository 创建角色包含关系仓储实例
func NewRoleInclusionRepository(roleInclusionDAO dao.RoleInclusionDAO) RoleInclusionRepository {
	return &roleInclusionRepository{
		roleInclusionDAO: roleInclusionDAO,
	}
}

func (r *roleInclusionRepository) Create(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error) {
	created, err := r.roleInclusionDAO.Create(ctx, r.toEntity(roleInclusion))
	if err != nil {
		return domain.RoleInclusion{}, err
	}
	return r.toDomain(created), nil
}

func (r *roleInclusionRepository) FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error) {
	roleInclusion, err := r.roleInclusionDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.RoleInclusion{}, err
	}
	return r.toDomain(roleInclusion), nil
}

func (r *roleInclusionRepository) FindByBizIDAndIncludingRoleIDs(ctx context.Context, bizID int64, includingRoleIDs []int64, offset, limit int) ([]domain.RoleInclusion, error) {
	roleInclusions, err := r.roleInclusionDAO.FindByBizIDAndIncludingRoleID(ctx, bizID, includingRoleIDs, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(roleInclusions, func(_ int, src dao.RoleInclusion) domain.RoleInclusion {
		return r.toDomain(src)
	}), nil
}

func (r *roleInclusionRepository) FindByBizIDAndIncludedRoleIDs(ctx context.Context, bizID int64, includedRoleIDs []int64, offset, limit int) ([]domain.RoleInclusion, error) {
	roleInclusions, err := r.roleInclusionDAO.FindByBizIDAndIncludedRoleID(ctx, bizID, includedRoleIDs, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(roleInclusions, func(_ int, src dao.RoleInclusion) domain.RoleInclusion {
		return r.toDomain(src)
	}), nil
}

func (r *roleInclusionRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.roleInclusionDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *roleInclusionRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, error) {
	roleInclusions, err := r.roleInclusionDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(roleInclusions, func(_ int, src dao.RoleInclusion) domain.RoleInclusion {
		return r.toDomain(src)
	}), nil
}

func (r *roleInclusionRepository) toEntity(ri domain.RoleInclusion) dao.RoleInclusion {
	return dao.RoleInclusion{
		ID:                ri.ID,
		BizID:             ri.BizID,
		IncludingRoleID:   ri.IncludingRole.ID,
		IncludingRoleType: ri.IncludingRole.Type,
		IncludingRoleName: ri.IncludingRole.Name,
		IncludedRoleID:    ri.IncludedRole.ID,
		IncludedRoleType:  ri.IncludedRole.Type,
		IncludedRoleName:  ri.IncludedRole.Name,
		Ctime:             ri.Ctime,
		Utime:             ri.Utime,
	}
}

func (r *roleInclusionRepository) toDomain(ri dao.RoleInclusion) domain.RoleInclusion {
	return domain.RoleInclusion{
		ID:    ri.ID,
		BizID: ri.BizID,
		IncludingRole: domain.Role{
			ID:   ri.IncludingRoleID,
			Type: ri.IncludingRoleType,
			Name: ri.IncludingRoleName,
		},
		IncludedRole: domain.Role{
			ID:   ri.IncludedRoleID,
			Type: ri.IncludedRoleType,
			Name: ri.IncludedRoleName,
		},
		Ctime: ri.Ctime,
		Utime: ri.Utime,
	}
}
