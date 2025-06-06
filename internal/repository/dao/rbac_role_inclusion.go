package dao

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

// RoleInclusion 角色包含关系表
type RoleInclusion struct {
	ID                int64  `gorm:"primaryKey;autoIncrement;comment:角色包含关系ID'"`
	BizID             int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:1;index:idx_biz_including_role,priority:1;index:idx_biz_included_role,priority:1;comment:'业务ID'"`
	IncludingRoleID   int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:2;index:idx_biz_including_role,priority:2;comment:'包含者角色ID（拥有其他角色权限）'"`
	IncludingRoleType string `gorm:"type:VARCHAR(255);NOT NULL;comment:'包含者角色类型（冗余字段，加速查询）'"`
	IncludingRoleName string `gorm:"type:VARCHAR(255);NOT NULL;comment:'包含者角色名称（冗余字段，加速查询）'"`
	IncludedRoleID    int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:uk_biz_including_included,priority:3;index:idx_biz_included_role,priority:2;comment:'被包含角色ID（权限被包含）'"`
	IncludedRoleType  string `gorm:"type:VARCHAR(255);NOT NULL;comment:'被包含角色类型（冗余字段，加速查询）'"`
	IncludedRoleName  string `gorm:"type:VARCHAR(255);NOT NULL;comment:'被包含角色名称（冗余字段，加速查询）'"`
	Ctime             int64
	Utime             int64
}

func (RoleInclusion) TableName() string {
	return "role_inclusions"
}

// RoleInclusionDAO 角色包含关系数据访问接口
type RoleInclusionDAO interface {
	Create(ctx context.Context, roleInclusion RoleInclusion) (RoleInclusion, error)

	FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]RoleInclusion, error)
	FindByBizIDAndID(ctx context.Context, bizID, id int64) (RoleInclusion, error)
	FindByBizIDAndIncludingRoleIDs(ctx context.Context, bizID int64, includingRoleIDs []int64) ([]RoleInclusion, error)
	FindByBizIDAndIncludedRoleIDs(ctx context.Context, bizID int64, includedRoleIDs []int64) ([]RoleInclusion, error)

	DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error
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

func (r *roleInclusionDAO) Create(ctx context.Context, roleInclusion RoleInclusion) (RoleInclusion, error) {
	now := time.Now().UnixMilli()
	roleInclusion.Ctime = now
	roleInclusion.Utime = now
	err := r.db.WithContext(ctx).Create(&roleInclusion).Error
	return roleInclusion, err
}

func (r *roleInclusionDAO) FindByBizID(ctx context.Context, bizID int64, offset, limit int) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).Where("biz_id = ?", bizID).Offset(offset).Limit(limit).Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) FindByBizIDAndID(ctx context.Context, bizID, id int64) (RoleInclusion, error) {
	var roleInclusion RoleInclusion
	err := r.db.WithContext(ctx).Where("biz_id = ? AND id = ?", bizID, id).First(&roleInclusion).Error
	return roleInclusion, err
}

func (r *roleInclusionDAO) FindByBizIDAndIncludingRoleIDs(ctx context.Context, bizID int64, includingRoleIDs []int64) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).Where("biz_id = ? AND including_role_id IN (?)", bizID, includingRoleIDs).Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) FindByBizIDAndIncludedRoleIDs(ctx context.Context, bizID int64, includedRoleIDs []int64) ([]RoleInclusion, error) {
	var roleInclusions []RoleInclusion
	err := r.db.WithContext(ctx).Where("biz_id = ? AND included_role_id IN (?)", bizID, includedRoleIDs).Find(&roleInclusions).Error
	return roleInclusions, err
}

func (r *roleInclusionDAO) DeleteByBizIDAndID(ctx context.Context, bizID, id int64) error {
	return r.db.WithContext(ctx).
		Where("biz_id = ? AND id = ?", bizID, id).
		Delete(&RoleInclusion{}).Error
}
