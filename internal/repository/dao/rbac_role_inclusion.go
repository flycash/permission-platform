package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

// RoleInclusion 角色包含关系表
type RoleInclusion struct {
	ID                int64    `gorm:"primaryKey;autoIncrement;comment:角色包含关系ID'"`
	BizID             int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:1;index:idx_biz_including_role,priority:1;index:idx_biz_included_role,priority:1;comment:'业务ID'"`
	IncludingRoleID   int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:2;index:idx_biz_including_role,priority:2;comment:'包含者角色ID（拥有其他角色权限）'"`
	IncludingRoleType RoleType `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;comment:'包含者角色类型（冗余字段，加速查询）'"`
	IncludingRoleName string   `gorm:"type:VARCHAR(100);NOT NULL;comment:'包含者角色名称（冗余字段，加速查询）'"`
	IncludedRoleID    int64    `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:3;index:idx_biz_included_role,priority:2;comment:'被包含角色ID（权限被包含）'"`
	IncludedRoleType  RoleType `gorm:"type:ENUM('system', 'custom', 'temporary');NOT NULL;comment:'被包含角色类型（冗余字段，加速查询）'"`
	IncludedRoleName  string   `gorm:"type:VARCHAR(100);NOT NULL;comment:'被包含角色名称（冗余字段，加速查询）'"`
	Ctime             int64
	Utime             int64
}

func (RoleInclusion) TableName() string {
	return "role_inclusions"
}

// RoleInclusionDAO 角色包含关系数据访问接口
type RoleInclusionDAO interface {
	// FindByIncludingRoleID 查找特定角色包含的所有角色
	FindByIncludingRoleID(ctx context.Context, bizID int64, includingRoleID int64) ([]RoleInclusion, error)

	// GetByID 根据ID获取角色包含关系
	GetByID(ctx context.Context, id int64) (RoleInclusion, error)

	// Create 创建角色包含关系
	Create(ctx context.Context, roleInclusion RoleInclusion) (RoleInclusion, error)

	// Delete 删除角色包含关系
	Delete(ctx context.Context, bizID, includingRoleID, includedRoleID int64) error

	// FindByBizIDAndRoleID 根据业务ID和角色ID查找角色包含关系
	FindByBizIDAndRoleID(ctx context.Context, bizID, roleID int64) ([]RoleInclusion, error)
}

// roleInclusionDAO 角色包含关系数据访问实现
type roleInclusionDAO struct {
	db *egorm.Component
}

// NewRoleInclusionDAO 创建角色包含关系数据访问对象
func NewRoleInclusionDAO(db *egorm.Component) RoleInclusionDAO {
	return &roleInclusionDAO{
		db: db,
	}
}

func (r *roleInclusionDAO) FindByIncludingRoleID(ctx context.Context, bizID, includingRoleID int64) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).Where("biz_id = ? AND including_role_id = ?", bizID, includingRoleID).Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) GetByID(ctx context.Context, id int64) (RoleInclusion, error) {
	var roleInclusion RoleInclusion
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&roleInclusion).Error
	return roleInclusion, err
}

func (r *roleInclusionDAO) Create(ctx context.Context, roleInclusion RoleInclusion) (RoleInclusion, error) {
	now := time.Now().UnixMilli()
	roleInclusion.Ctime = now
	roleInclusion.Utime = now
	err := r.db.WithContext(ctx).Create(&roleInclusion).Error
	return roleInclusion, err
}

func (r *roleInclusionDAO) Delete(ctx context.Context, bizID, includingRoleID, includedRoleID int64) error {
	return r.db.WithContext(ctx).
		Where("biz_id = ? AND including_role_id = ? AND included_role_id = ?", bizID, includingRoleID, includedRoleID).
		Delete(&RoleInclusion{}).Error
}

func (r *roleInclusionDAO) FindByBizIDAndRoleID(ctx context.Context, bizID, roleID int64) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND (including_role_id = ? OR included_role_id = ?)", bizID, roleID, roleID).
		Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) GetByIDs(ctx context.Context, ids []int64) (map[int64]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).Where("id IN (?)", ids).Find(&roleInclusions).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]RoleInclusion, len(roleInclusions))
	for i := range roleInclusions {
		result[roleInclusions[i].ID] = roleInclusions[i]
	}
	return result, nil
}

func (r *roleInclusionDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) FindByIncludedRoleID(ctx context.Context, includedRoleID int64) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).Where("included_role_id = ?", includedRoleID).Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) FindByBizIDAndIncludingRoleType(ctx context.Context, bizID int64, roleType RoleType, offset, limit int) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND including_role_type = ?", bizID, roleType).
		Offset(offset).
		Limit(limit).
		Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) FindByBizIDAndIncludedRoleType(ctx context.Context, bizID int64, roleType RoleType, offset, limit int) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).
		Where("biz_id = ? AND included_role_type = ?", bizID, roleType).
		Offset(offset).
		Limit(limit).
		Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) ExistsByIncludingAndIncludedRoleID(ctx context.Context, bizID, includingRoleID, includedRoleID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&RoleInclusion{}).
		Where("biz_id = ? AND including_role_id = ? AND included_role_id = ?", bizID, includingRoleID, includedRoleID).
		Count(&count).Error
	return count > 0, err
}

func (r *roleInclusionDAO) BatchCreate(ctx context.Context, roleInclusions []RoleInclusion) error {
	if len(roleInclusions) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	for i := range roleInclusions {
		roleInclusions[i].Ctime = now
		roleInclusions[i].Utime = now
	}

	return r.db.WithContext(ctx).CreateInBatches(roleInclusions, batchSize).Error
}

func (r *roleInclusionDAO) DeleteByIncludingAndIncludedRoleID(ctx context.Context, bizID, includingRoleID, includedRoleID int64) error {
	return r.db.WithContext(ctx).
		Where("biz_id = ? AND including_role_id = ? AND included_role_id = ?", bizID, includingRoleID, includedRoleID).
		Delete(&RoleInclusion{}).Error
}

func (r *roleInclusionDAO) DeleteByIncludingRoleID(ctx context.Context, includingRoleID int64) error {
	return r.db.WithContext(ctx).Where("including_role_id = ?", includingRoleID).Delete(&RoleInclusion{}).Error
}

func (r *roleInclusionDAO) DeleteByIncludedRoleID(ctx context.Context, includedRoleID int64) error {
	return r.db.WithContext(ctx).Where("included_role_id = ?", includedRoleID).Delete(&RoleInclusion{}).Error
}
