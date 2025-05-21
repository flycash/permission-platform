package dao

import (
	"context"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"

	"github.com/pkg/errors"

	"github.com/ego-component/egorm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrUpdateFailed = errors.New("update failed")

// Policy 策略表模型
type Policy struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement;"`
	BizID       int64  `gorm:"column:biz_id;index:idx_biz_id;comment:业务ID"`
	Name        string `gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx_biz_name;comment:策略名称" json:"name"`
	Description string `gorm:"column:description;type:text;comment:策略描述" json:"description"`
	Status      string `gorm:"column:status;type:enum('active','inactive');not null;default:active;index:idx_status;comment:策略状态" json:"status"`
	ExecuteType string `gorm:"column:execute_type;type:varchar(255);default:logic"`
	Ctime       int64  `gorm:"column:ctime;comment:创建时间"`
	Utime       int64  `gorm:"column:utime;comment:更新时间"`
}

// TableName 指定表名
func (p Policy) TableName() string {
	return "policies"
}

// PolicyRule 策略规则表模型
type PolicyRule struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement;"`
	BizID     int64  `gorm:"column:biz_id;index:idx_biz_id;comment:业务ID"`
	PolicyID  int64  `gorm:"column:policy_id;not null;index:idx_policy_id;comment:策略ID"`
	AttrDefID int64  `gorm:"column:attr_def_id;not null;index:idx_attribute_id;comment:属性定义ID"`
	Value     string `gorm:"column:value;type:text;comment:比较值，取决于类型"`
	Left      int64  `gorm:"column:left;comment:左规则ID"`
	Right     int64  `gorm:"column:right;comment:右规则ID"`
	Operator  string `gorm:"column:operator;type:varchar(255);not null;comment:操作符"`
	Ctime     int64  `gorm:"column:ctime;comment:创建时间"`
	Utime     int64  `gorm:"column:utime;comment:更新时间"`
}

// TableName 指定表名
func (r PolicyRule) TableName() string {
	return "policy_rules"
}

// PermissionPolicy 权限策略关联表模型
type PermissionPolicy struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement;"`
	BizID        int64  `gorm:"column:biz_id;index:idx_biz_id;comment:业务ID;uniqueIndex:idx_permission_policy_bizId"`
	Effect       string `gorm:"column:effect;type:varchar(50)"`
	PermissionID int64  `gorm:"column:permission_id;not null;uniqueIndex:idx_permission_policy_bizId;index:idx_permission_id;comment:权限ID"`
	PolicyID     int64  `gorm:"column:policy_id;not null;uniqueIndex:idx_permission_policy_bizId;index:idx_policy_id;comment:策略ID"`
	Ctime        int64  `gorm:"column:ctime;comment:创建时间"`
	Utime        int64  `gorm:"column:utime;comment:创建时间"`
}

// TableName 指定表名
func (p PermissionPolicy) TableName() string {
	return "permission_policies"
}

// -------- Policy DAO Interface --------

// PolicyDAO 综合策略数据访问接口
type PolicyDAO interface {
	// Policy 相关方法
	SavePolicy(ctx context.Context, policy Policy) (int64, error)
	UpdatePolicyStatus(ctx context.Context, id int64, status string) error
	DeletePolicy(ctx context.Context, bizID, id int64) error
	FindPolicy(ctx context.Context, bizID, id int64) (Policy, error)
	FindPoliciesByIDs(ctx context.Context, ids []int64) ([]Policy, error)
	FindPoliciesByBiz(ctx context.Context, bizID int64) ([]Policy, error)
	PolicyList(ctx context.Context, bizID int64, offset, limit int) ([]Policy, error)
	PolicyListCount(ctx context.Context, bizID int64) (int64, error)

	// PolicyRule 相关方法
	SavePolicyRule(ctx context.Context, rule PolicyRule) (int64, error)
	DeletePolicyRule(ctx context.Context, bizID, id int64) error
	FindPolicyRule(ctx context.Context, id int64) (PolicyRule, error)
	FindPolicyRulesByPolicyID(ctx context.Context, bizID, policyID int64) ([]PolicyRule, error)
	FindPolicyRulesByPolicyIDs(ctx context.Context, policyIDs []int64) (map[int64][]PolicyRule, error)

	// PermissionPolicy 相关方法
	SavePermissionPolicy(ctx context.Context, permissionPolicy PermissionPolicy) error
	DeletePermissionPolicy(ctx context.Context, bizID int64, permissionID int64, policyID int64) error
	FindPoliciesByPermission(ctx context.Context, bizID int64, permissionIDs []int64) ([]PermissionPolicy, error)
}

type policyDAO struct {
	db *egorm.Component
}

func (p *policyDAO) PolicyList(ctx context.Context, bizID int64, offset, limit int) ([]Policy, error) {
	var list []Policy
	err := p.db.WithContext(ctx).Where("biz_id = ?", bizID).
		Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}

func (p *policyDAO) PolicyListCount(ctx context.Context, bizID int64) (int64, error) {
	var count int64
	err := p.db.WithContext(ctx).
		Model(&Policy{}).
		Where("biz_id = ?", bizID).Count(&count).Error
	return count, err
}

// NewPolicyDAO 创建综合策略数据访问对象
func NewPolicyDAO(db *egorm.Component) PolicyDAO {
	return &policyDAO{db: db}
}

func (p *policyDAO) SavePolicy(ctx context.Context, policy Policy) (int64, error) {
	now := time.Now().UnixMilli()
	if policy.ID == 0 {
		policy.Ctime = now
	}
	policy.Utime = now

	err := p.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"description", "utime"}),
		}).Create(&policy).Error
	return policy.ID, err
}

func (p *policyDAO) UpdatePolicyStatus(ctx context.Context, id int64, status string) error {
	res := p.db.WithContext(ctx).
		Model(&Policy{}).
		Where("id = ? ", id).
		Updates(map[string]interface{}{"status": status, "utime": time.Now().UnixMilli()})
	if res.RowsAffected == 0 {
		return ErrUpdateFailed
	}
	return res.Error
}

func (p *policyDAO) DeletePolicy(ctx context.Context, bizID, id int64) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除策略与权限的关联关系
		if err := tx.Where("biz_id = ? AND policy_id = ?", bizID, id).Delete(&PermissionPolicy{}).Error; err != nil {
			return err
		}

		// 2. 删除策略关联的规则
		if err := tx.Where("biz_id = ? AND policy_id = ?", bizID, id).Delete(&PolicyRule{}).Error; err != nil {
			return err
		}

		// 3. 删除策略本身
		return tx.Where("id = ? AND biz_id = ?", id, bizID).Delete(&Policy{}).Error
	})
}

func (p *policyDAO) FindPolicy(ctx context.Context, bizID, id int64) (Policy, error) {
	var policy Policy
	err := p.db.WithContext(ctx).
		Where("id = ? AND biz_id = ?", id, bizID).
		First(&policy).Error
	return policy, err
}

func (p *policyDAO) FindPoliciesByIDs(ctx context.Context, ids []int64) ([]Policy, error) {
	var policies []Policy
	err := p.db.WithContext(ctx).
		Where(" id IN ? AND status = ?", ids, domain.PolicyStatusActive).
		Find(&policies).Error
	return policies, err
}

func (p *policyDAO) FindPoliciesByBiz(ctx context.Context, bizID int64) ([]Policy, error) {
	var policies []Policy
	err := p.db.WithContext(ctx).
		Where("biz_id = ?", bizID).
		Find(&policies).Error
	return policies, err
}

func (p *policyDAO) SavePolicyRule(ctx context.Context, rule PolicyRule) (int64, error) {
	now := time.Now().UnixMilli()
	if rule.ID == 0 {
		rule.Ctime = now
	}
	rule.Utime = now

	err := p.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "left", "right", "operator", "utime"}),
		}).Create(&rule).Error
	return rule.ID, err
}

func (p *policyDAO) DeletePolicyRule(ctx context.Context, bizID, id int64) error {
	return p.db.WithContext(ctx).
		Where("id = ? AND biz_id", id, bizID).
		Delete(&PolicyRule{}).Error
}

func (p *policyDAO) FindPolicyRule(ctx context.Context, id int64) (PolicyRule, error) {
	var rule PolicyRule
	err := p.db.WithContext(ctx).
		Where("id = ?", id).
		First(&rule).Error
	return rule, err
}

func (p *policyDAO) FindPolicyRulesByPolicyID(ctx context.Context, bizID, policyID int64) ([]PolicyRule, error) {
	var rules []PolicyRule
	err := p.db.WithContext(ctx).
		Where(" policy_id = ? AND biz_id = ?", policyID, bizID).
		Find(&rules).Error
	return rules, err
}

func (p *policyDAO) FindPolicyRulesByPolicyIDs(ctx context.Context, policyIDs []int64) (map[int64][]PolicyRule, error) {
	var rules []PolicyRule
	err := p.db.WithContext(ctx).
		Where(" policy_id IN ?", policyIDs).
		Find(&rules).Error
	if err != nil {
		return nil, err
	}

	// 将规则按策略ID分组
	result := make(map[int64][]PolicyRule)
	for _, rule := range rules {
		result[rule.PolicyID] = append(result[rule.PolicyID], rule)
	}

	return result, nil
}

func (p *policyDAO) SavePermissionPolicy(ctx context.Context, permissionPolicy PermissionPolicy) error {
	permissionPolicy.Ctime = time.Now().UnixMilli()
	permissionPolicy.Utime = time.Now().UnixMilli()
	err := p.db.WithContext(ctx).Create(&permissionPolicy).Error
	return err
}

func (p *policyDAO) DeletePermissionPolicy(ctx context.Context, bizID, permissionID, policyID int64) error {
	return p.db.WithContext(ctx).
		Where("biz_id = ? AND permission_id = ? AND policy_id = ?", bizID, permissionID, policyID).
		Delete(&PermissionPolicy{}).Error
}

func (p *policyDAO) FindPoliciesByPermission(ctx context.Context, bizID int64, permissionIDs []int64) ([]PermissionPolicy, error) {
	var relations []PermissionPolicy
	err := p.db.WithContext(ctx).
		Where("biz_id = ? AND permission_id in ?", bizID, permissionIDs).
		Find(&relations).Error
	return relations, err
}
