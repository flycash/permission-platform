package repository

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/ecodeclub/ekit/mapx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

var _ RoleInclusionRepository = (*RoleInclusionReloadCacheRepository)(nil)

// RoleInclusionReloadCacheRepository 角色包含关系仓储实现
type RoleInclusionReloadCacheRepository struct {
	repo          *RoleInclusionDefaultRepository
	userRoleRepo  *UserRoleDefaultRepository
	cacheReloader UserPermissionCacheReloader
	logger        *elog.Component
}

// NewRoleInclusionReloadCacheRepository 创建可以重载缓存的角色包含关系仓储实例
func NewRoleInclusionReloadCacheRepository(
	repo *RoleInclusionDefaultRepository,
	userRoleRepo *UserRoleDefaultRepository,
	cacheReloader UserPermissionCacheReloader,
) *RoleInclusionReloadCacheRepository {
	return &RoleInclusionReloadCacheRepository{
		repo:          repo,
		userRoleRepo:  userRoleRepo,
		cacheReloader: cacheReloader,
		logger:        elog.DefaultLogger.With(elog.FieldName("RoleInclusionReloadCacheRepository")),
	}
}

func (r *RoleInclusionReloadCacheRepository) Create(ctx context.Context, roleInclusion domain.RoleInclusion) (domain.RoleInclusion, error) {
	created, err := r.repo.Create(ctx, roleInclusion)
	if err != nil {
		return domain.RoleInclusion{}, err
	}
	if err1 := r.cacheReloader.Reload(ctx, r.getAffectedUsers(ctx, created.BizID, created.IncludingRole.ID)); err1 != nil {
		r.logger.Warn("创建角色包含关系成功后，重新加载所有受影响用户的缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", created.BizID),
			elog.Any("includingRoleID", created.IncludingRole.ID),
			elog.Any("includedRoleID", created.IncludedRole.ID),
		)
	}
	return created, nil
}

func (r *RoleInclusionReloadCacheRepository) getAffectedUsers(ctx context.Context, bizID, includedRoleID int64) []domain.User {
	roleIDs, err := r.getAffectedRoleIDs(ctx, bizID, includedRoleID)
	if err != nil {
		return nil
	}
	userRoles, err := r.userRoleRepo.FindByBizIDAndRoleIDs(ctx, bizID, roleIDs)
	if err != nil {
		return nil
	}
	return slice.Map(userRoles, func(_ int, src domain.UserRole) domain.User {
		return domain.User{
			ID:    src.ID,
			BizID: src.BizID,
		}
	})
}

func (r *RoleInclusionReloadCacheRepository) getAffectedRoleIDs(ctx context.Context, bizID, includedRoleID int64) ([]int64, error) {
	allRoleIDs := make(map[int64]any)
	allRoleIDs[includedRoleID] = struct{}{}

	includedIDs := []int64{includedRoleID}
	for {
		inclusions, err := r.repo.FindByBizIDAndIncludedRoleIDs(ctx, bizID, includedIDs)
		if err != nil {
			return nil, err
		}
		if len(inclusions) == 0 {
			break
		}
		// 沿着包含关系链，逆向查找
		// A->B->C, 当C添加了权限，此时IncludedRoleID=C，然后要沿着与之关联的 IncludingRoleID 逆向查找 —— 找到B，再找到A
		includedIDs = slice.Map(inclusions, func(_ int, src domain.RoleInclusion) int64 {
			allRoleIDs[src.IncludingRole.ID] = struct{}{}
			return src.IncludingRole.ID
		})
	}
	return mapx.Keys(allRoleIDs), nil
}

func (r *RoleInclusionReloadCacheRepository) FindByBizIDAndID(ctx context.Context, bizID, id int64) (domain.RoleInclusion, error) {
	return r.repo.FindByBizIDAndID(ctx, bizID, id)
}

func (r *RoleInclusionReloadCacheRepository) FindByBizIDAndIncludingRoleIDs(ctx context.Context, bizID int64, includingRoleIDs []int64) ([]domain.RoleInclusion, error) {
	return r.repo.FindByBizIDAndIncludingRoleIDs(ctx, bizID, includingRoleIDs)
}

func (r *RoleInclusionReloadCacheRepository) FindByBizIDAndIncludedRoleIDs(ctx context.Context, bizID int64, includedRoleIDs []int64) ([]domain.RoleInclusion, error) {
	return r.repo.FindByBizIDAndIncludedRoleIDs(ctx, bizID, includedRoleIDs)
}

func (r *RoleInclusionReloadCacheRepository) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	deleted, err := r.repo.FindByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	err = r.repo.DeleteByBizIDAndID(ctx, bizID, id)
	if err != nil {
		return err
	}
	if err1 := r.cacheReloader.Reload(ctx, r.getAffectedUsers(ctx, deleted.BizID, deleted.IncludingRole.ID)); err1 != nil {
		r.logger.Warn("删除角色包含关系成功后，重新加载所有受影响用户的缓存失败",
			elog.FieldErr(err1),
			elog.Any("bizID", deleted.BizID),
			elog.Any("includingRoleID", deleted.IncludingRole.ID),
			elog.Any("includedRoleID", deleted.IncludedRole.ID),
		)
	}
	return nil
}

func (r *RoleInclusionReloadCacheRepository) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]domain.RoleInclusion, error) {
	return r.repo.FindByBizID(ctx, bizID, offset, limit)
}
