package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

var _ RolePermissionRepository = (*RolePermissionDefaultRepository)(nil)

// RolePermissionRepository 角色权限关系仓储接口
type RolePermissionRepository interface {
	Create(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error)

	FindByBizID(ctx context.Context, bizID int64) ([]domain.RolePermission, error)
	FindByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64) ([]domain.RolePermission, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
}

// RolePermissionDefaultRepository 角色权限关系仓储实现
type RolePermissionDefaultRepository struct {
	rolePermissionDAO dao.RolePermissionDAO
	logger            *elog.Component
}

// NewRolePermissionDefaultRepository 创建角色权限关系仓储实例
func NewRolePermissionDefaultRepository(rolePermissionDAO dao.RolePermissionDAO) *RolePermissionDefaultRepository {
	return &RolePermissionDefaultRepository{
		rolePermissionDAO: rolePermissionDAO,
		logger:            elog.DefaultLogger,
	}
}

func (r *RolePermissionDefaultRepository) Create(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error) {
	created, err := r.rolePermissionDAO.Create(ctx, r.toEntity(rolePermission))
	if err != nil {
		r.logger.Info("为角色添加权限失败",
			elog.FieldErr(err),
			elog.Any("rolePermission", rolePermission),
			elog.Int64("roleId", rolePermission.Role.ID),
			elog.Int64("permissionId", rolePermission.Permission.ID),
			elog.Int64("bizID", rolePermission.BizID),
		)
		return domain.RolePermission{}, err
	} else {
		r.logger.Info("为角色添加权限",
			elog.Any("rolePermission", rolePermission),
			elog.Int64("roleId", rolePermission.Role.ID),
			elog.Int64("permissionId", rolePermission.Permission.ID),
			elog.Int64("bizID", rolePermission.BizID),
		)
	}
	return r.toDomain(created), nil
}

func (r *RolePermissionDefaultRepository) FindByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64) ([]domain.RolePermission, error) {
	rolePermissions, err := r.rolePermissionDAO.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs)
	if err != nil {
		return nil, err
	}

	return slice.Map(rolePermissions, func(_ int, src dao.RolePermission) domain.RolePermission {
		return r.toDomain(src)
	}), nil
}

func (r *RolePermissionDefaultRepository) FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.RolePermission, error) {
	rp, err := r.rolePermissionDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.RolePermission{}, err
	}
	return r.toDomain(rp), nil
}

func (r *RolePermissionDefaultRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	err := r.rolePermissionDAO.DeleteByBizIDAndID(ctx, bizID, id)
	if err != nil {
		r.logger.Error("为角色删除权限失败",
			elog.FieldErr(err),
			elog.Int64("bizID", bizID),
			elog.Any("角色,权限关联id", id))
	} else {
		r.logger.Info("为角色删除权限",
			elog.Int64("bizID", bizID),
			elog.Any("角色,权限关联id", id),
		)
	}
	return err
}

func (r *RolePermissionDefaultRepository) FindByBizID(ctx context.Context, bizID int64) ([]domain.RolePermission, error) {
	rolePermissions, err := r.rolePermissionDAO.FindByBizID(ctx, bizID)
	if err != nil {
		return nil, err
	}

	return slice.Map(rolePermissions, func(_ int, src dao.RolePermission) domain.RolePermission {
		return r.toDomain(src)
	}), nil
}

func (r *RolePermissionDefaultRepository) toEntity(rp domain.RolePermission) dao.RolePermission {
	return dao.RolePermission{
		ID:               rp.ID,
		BizID:            rp.BizID,
		RoleID:           rp.Role.ID,
		RoleName:         rp.Role.Name,
		RoleType:         rp.Role.Type,
		PermissionID:     rp.Permission.ID,
		ResourceType:     rp.Permission.Resource.Type,
		ResourceKey:      rp.Permission.Resource.Key,
		PermissionAction: rp.Permission.Action,
		Ctime:            rp.Ctime,
		Utime:            rp.Utime,
	}
}

func (r *RolePermissionDefaultRepository) toDomain(rp dao.RolePermission) domain.RolePermission {
	return domain.RolePermission{
		ID:    rp.ID,
		BizID: rp.BizID,
		Role: domain.Role{
			ID:   rp.RoleID,
			Name: rp.RoleName,
			Type: rp.RoleType,
		},
		Permission: domain.Permission{
			ID: rp.PermissionID,
			Resource: domain.Resource{
				Type: rp.ResourceType,
				Key:  rp.ResourceKey,
			},
			Action: rp.PermissionAction,
		},
		Ctime: rp.Ctime,
		Utime: rp.Utime,
	}
}
