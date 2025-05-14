package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// RoleRepository 角色仓储接口
type RoleRepository interface {
	Create(ctx context.Context, role domain.Role) (domain.Role, error)
	FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Role, error)
	FindByBizIDAndType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, error)
	UpdateByBizIDAndID(ctx context.Context, role domain.Role) (domain.Role, error)
	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, error)
}

// roleRepository 角色仓储实现
type roleRepository struct {
	roleDAO dao.RoleDAO
}

// NewRoleRepository 创建角色仓储实例
func NewRoleRepository(roleDAO dao.RoleDAO) RoleRepository {
	return &roleRepository{
		roleDAO: roleDAO,
	}
}

func (r *roleRepository) Create(ctx context.Context, role domain.Role) (domain.Role, error) {
	created, err := r.roleDAO.Create(ctx, r.toEntity(role))
	if err != nil {
		return domain.Role{}, err
	}
	return r.toDomain(created), nil
}

func (r *roleRepository) FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.Role, error) {
	role, err := r.roleDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.Role{}, err
	}
	return r.toDomain(role), nil
}

func (r *roleRepository) FindByBizIDAndType(ctx context.Context, bizID int64, roleType string, offset, limit int) ([]domain.Role, error) {
	roles, err := r.roleDAO.FindByBizIDAndType(ctx, bizID, roleType, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(roles, func(_ int, src dao.Role) domain.Role {
		return r.toDomain(src)
	}), nil
}

func (r *roleRepository) UpdateByBizIDAndID(ctx context.Context, role domain.Role) (domain.Role, error) {
	err := r.roleDAO.UpdateByBizIDAndID(ctx, r.toEntity(role))
	if err != nil {
		return domain.Role{}, err
	}
	return role, nil
}

func (r *roleRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.roleDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *roleRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.Role, error) {
	roles, err := r.roleDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(roles, func(_ int, src dao.Role) domain.Role {
		return r.toDomain(src)
	}), nil
}

func (r *roleRepository) toEntity(role domain.Role) dao.Role {
	return dao.Role{
		ID:          role.ID,
		BizID:       role.BizID,
		Type:        role.Type,
		Name:        role.Name,
		Description: role.Description,
		Metadata:    role.Metadata,
		Ctime:       role.Ctime,
		Utime:       role.Utime,
	}
}

func (r *roleRepository) toDomain(role dao.Role) domain.Role {
	return domain.Role{
		ID:          role.ID,
		BizID:       role.BizID,
		Type:        role.Type,
		Name:        role.Name,
		Description: role.Description,
		Metadata:    role.Metadata,
		Ctime:       role.Ctime,
		Utime:       role.Utime,
	}
}
