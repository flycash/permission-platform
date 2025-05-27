package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

var _ UserRoleRepository = (*UserRoleDefaultRepository)(nil)

// UserRoleRepository 用户角色关系仓储接口
type UserRoleRepository interface {
	Create(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error)

	FindByBizID(ctx context.Context, bizID int64) ([]domain.UserRole, error)
	FindByBizIDAndUserID(ctx context.Context, bizID, userID int64) ([]domain.UserRole, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
}

// UserRoleDefaultRepository 用户角色关系仓储实现
type UserRoleDefaultRepository struct {
	userRoleDAO dao.UserRoleDAO
	logger      *elog.Component
}

// NewUserRoleDefaultRepository 创建用户角色关系仓储实例
func NewUserRoleDefaultRepository(userRoleDAO dao.UserRoleDAO) *UserRoleDefaultRepository {
	return &UserRoleDefaultRepository{
		userRoleDAO: userRoleDAO,
		logger:      elog.DefaultLogger,
	}
}

func (r *UserRoleDefaultRepository) Create(ctx context.Context, userRole domain.UserRole) (domain.UserRole, error) {
	created, err := r.userRoleDAO.Create(ctx, r.toEntity(userRole))
	if err != nil {
		r.logger.Error("授予角色权限失败",
			elog.Int64("bizId", userRole.BizID),
			elog.Int64("userId", userRole.UserID),
			elog.Int64("roleId", userRole.Role.ID),
			elog.String("roleName", userRole.Role.Name),
			elog.FieldErr(err),
		)
		return domain.UserRole{}, err
	} else {
		r.logger.Info("授予角色权限",
			elog.Int64("bizId", userRole.BizID),
			elog.Int64("userId", userRole.UserID),
			elog.Int64("roleId", userRole.Role.ID),
			elog.String("roleName", userRole.Role.Name),
			elog.Any("userRole", created),
		)
	}
	return r.toDomain(created), nil
}

func (r *UserRoleDefaultRepository) FindByBizID(ctx context.Context, bizID int64) ([]domain.UserRole, error) {
	userRoles, err := r.userRoleDAO.FindByBizID(ctx, bizID)
	if err != nil {
		return nil, err
	}

	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toDomain(src)
	}), nil
}

func (r *UserRoleDefaultRepository) FindByBizIDAndUserID(ctx context.Context, bizID, userID int64) ([]domain.UserRole, error) {
	userRoles, err := r.userRoleDAO.FindByBizIDAndUserID(ctx, bizID, userID)
	if err != nil {
		return nil, err
	}

	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toDomain(src)
	}), nil
}

func (r *UserRoleDefaultRepository) FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.UserRole, error) {
	ur, err := r.userRoleDAO.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return domain.UserRole{}, err
	}
	return r.toDomain(ur), nil
}

func (r *UserRoleDefaultRepository) FindByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64) ([]domain.UserRole, error) {
	userRoles, err := r.userRoleDAO.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs)
	if err != nil {
		return nil, err
	}
	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.UserRole {
		return r.toDomain(src)
	}), nil
}

func (r *UserRoleDefaultRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	err := r.userRoleDAO.DeleteByBizIDAndID(ctx, bizID, id)
	if err != nil {
		r.logger.Error("撤销角色权限失败",
			elog.FieldErr(err),
			elog.Int64("bizId", bizID),
			elog.Int64("userRoleId", id),
		)
	} else {
		r.logger.Info("撤销角色权限",
			elog.Int64("bizId", bizID),
			elog.Int64("userRoleId", id),
		)
	}
	return err
}

func (r *UserRoleDefaultRepository) toEntity(ur domain.UserRole) dao.UserRole {
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

func (r *UserRoleDefaultRepository) toDomain(ur dao.UserRole) domain.UserRole {
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
