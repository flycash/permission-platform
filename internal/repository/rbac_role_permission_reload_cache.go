package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"github.com/ecodeclub/ekit/mapx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

var _ RolePermissionRepository = (*RolePermissionReloadCacheRepository)(nil)

// RolePermissionReloadCacheRepository 角色权限关系仓储实现
type RolePermissionReloadCacheRepository struct {
	repo             *RolePermissionDefaultRepository
	roleInclusionDAO dao.RoleInclusionDAO
	userRoleDAO      dao.UserRoleDAO
	cacheReloader    UserPermissionCacheReloader
	logger           *elog.Component
}

// NewRolePermissionReloadCacheRepository 创建可以重载缓存的角色权限关系仓储实例
func NewRolePermissionReloadCacheRepository(
	repo *RolePermissionDefaultRepository,
	roleInclusionDAO dao.RoleInclusionDAO,
	userRoleRepoDAO dao.UserRoleDAO,
	cacheReloader UserPermissionCacheReloader,
) *RolePermissionReloadCacheRepository {
	return &RolePermissionReloadCacheRepository{
		repo:             repo,
		roleInclusionDAO: roleInclusionDAO,
		userRoleDAO:      userRoleRepoDAO,
		cacheReloader:    cacheReloader,
		logger:           elog.DefaultLogger.With(elog.FieldName("RolePermissionReloadCacheRepository")),
	}
}

func (r *RolePermissionReloadCacheRepository) Create(ctx context.Context, rolePermission domain.RolePermission) (domain.RolePermission, error) {
	created, err := r.repo.Create(ctx, rolePermission)
	if err != nil {
		return domain.RolePermission{}, err
	}
	if err1 := r.cacheReloader.Reload(ctx, r.getAffectedUsers(ctx, created.BizID, created.Role.ID)); err1 != nil {
		r.logger.Warn("创建角色权限关联关系成功后，重新加载所有受影响用户的缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", created.BizID),
			elog.Any("roleID", created.Role.ID),
			elog.Any("permissionID", created.Permission.ID),
		)
	}
	return created, nil
}

func (r *RolePermissionReloadCacheRepository) getAffectedUsers(ctx context.Context, bizID, includedRoleID int64) []domain.User {
	roleIDs, err := r.getAffectedRoleIDs(ctx, bizID, includedRoleID)
	if err != nil {
		return nil
	}
	userRoles, err := r.userRoleDAO.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs)
	if err != nil {
		return nil
	}
	return slice.Map(userRoles, func(_ int, src dao.UserRole) domain.User {
		return domain.User{
			ID:    src.ID,
			BizID: src.BizID,
		}
	})
}

func (r *RolePermissionReloadCacheRepository) getAffectedRoleIDs(ctx context.Context, bizID, includedRoleID int64) ([]int64, error) {
	allRoleIDs := make(map[int64]any)
	allRoleIDs[includedRoleID] = struct{}{}

	includedIDs := []int64{includedRoleID}
	for {
		inclusions, err := r.roleInclusionDAO.FindByBizIDAndIncludedRoleIDs(ctx, bizID, includedIDs)
		if err != nil {
			return nil, err
		}
		if len(inclusions) == 0 {
			break
		}
		// 沿着包含关系链，逆向查找
		// A->B->C, 当C添加了权限，此时IncludedRoleID=C，然后要沿着与之关联的 IncludingRoleID 逆向查找 —— 找到B，再找到A
		includedIDs = slice.Map(inclusions, func(_ int, src dao.RoleInclusion) int64 {
			allRoleIDs[src.IncludingRoleID] = struct{}{}
			return src.IncludingRoleID
		})
	}
	return mapx.Keys(allRoleIDs), nil
}

func (r *RolePermissionReloadCacheRepository) FindByBizIDAndRoleIDs(ctx context.Context, bizID int64, roleIDs []int64) ([]domain.RolePermission, error) {
	return r.repo.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs)
}

func (r *RolePermissionReloadCacheRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	deleted, err := r.repo.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	err = r.repo.DeleteByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	if err1 := r.cacheReloader.Reload(ctx, r.getAffectedUsers(ctx, deleted.BizID, deleted.Role.ID)); err1 != nil {
		r.logger.Warn("删除角色权限关联关系成功后，重新加载所有受影响用户的缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", deleted.BizID),
			elog.Any("roleID", deleted.Role.ID),
			elog.Any("permissionID", deleted.Permission.ID),
		)
	}
	return nil
}

func (r *RolePermissionReloadCacheRepository) FindByBizID(ctx context.Context, bizID int64) ([]domain.RolePermission, error) {
	return r.repo.FindByBizID(ctx, bizID)
}
