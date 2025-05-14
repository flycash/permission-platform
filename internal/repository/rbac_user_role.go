package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// UserRoleRepository 用户角色关系仓储接口
type UserRoleRepository interface {
	Create(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error)
	FindByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, error)
	FindValidByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserRole, error)
	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserRole, error)
}

// userRoleRepository 用户角色关系仓储实现
type userRoleRepository struct {
	userRoleDAO dao.UserRoleDAO
}

// NewUserRoleRepository 创建用户角色关系仓储实例
func NewUserRoleRepository(userRoleDAO dao.UserRoleDAO) UserRoleRepository {
	return &userRoleRepository{
		userRoleDAO: userRoleDAO,
	}
}

func (r *userRoleRepository) Create(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error) {
	created, err := r.userRoleDAO.Create(ctx, r.toEntity(userRole))
	if err != nil {
		return domain.UserRole{}, err
	}
	return r.toDomain(created), nil
}

func (r *userRoleRepository) FindByBizIDAndUserID(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, error) {
	userRoles, err := r.userRoleDAO.FindByBizIDAndUserID(ctx, bizID, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toDomain(src)
	}), nil
}

func (r *userRoleRepository) FindValidByBizIDAndUserID(ctx context.Context, bizID, userID, currentTime int64, offset, limit int) ([]domain.UserRole, error) {
	userRoles, err := r.userRoleDAO.FindValidUserRolesWithBizID(ctx, bizID, userID, currentTime, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toDomain(src)
	}), nil
}

func (r *userRoleRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.userRoleDAO.DeleteByBizIDAndID(ctx, bizID, id)
}

func (r *userRoleRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.UserRole, error) {
	userRoles, err := r.userRoleDAO.FindByBizID(ctx, bizID, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toDomain(src)
	}), nil
}

func (r *userRoleRepository) toEntity(ur domain.UserRole) dao.UserRole {
	return dao.UserRole{
		ID:        ur.ID,
		BizID:     ur.BizID,
		UserID:    ur.UserID,
		RoleID:    ur.Role.ID,
		RoleName:  ur.Role.Name,
		RoleType:  ur.Role.Type,
		StartTime: ur.StartTime,
		EndTime:   ur.EndTime,
		Ctime:     ur.Ctime,
		Utime:     ur.Utime,
	}
}

func (r *userRoleRepository) toDomain(ur dao.UserRole) domain.UserRole {
	return domain.UserRole{
		ID:     ur.ID,
		BizID:  ur.BizID,
		UserID: ur.UserID,
		Role: domain.Role{
			ID:   ur.RoleID,
			Type: ur.RoleType,
			Name: ur.RoleName,
		},
		StartTime: ur.StartTime,
		EndTime:   ur.EndTime,
		Ctime:     ur.Ctime,
		Utime:     ur.Utime,
	}
}
