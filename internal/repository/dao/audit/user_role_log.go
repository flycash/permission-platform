package audit

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

type UserRoleLog struct {
	ID           int64  `gorm:"primaryKey;autoIncrement;comment:'用户权限表变更日志表自增ID'"`
	Operation    string `gorm:"type:VARCHAR(255);NOT NULL;comment:'操作类型：INSERT/DELETE'"`
	BizID        int64  `gorm:"type:BIGINT;NOT NULL;comment:'业务ID'"`
	UserID       int64  `gorm:"type:BIGINT;NOT NULL;comment:'用户ID'"`
	BeforeRoleID int64  `gorm:"type:BIGINT;NOT NULL;DEFAULT 0;comment:'变更前的角色ID，Operation=INSERT 的情况下无意义'"`
	AfterRoleID  int64  `gorm:"type:BIGINT;NOT NULL;DEFAULT 0;comment:'变更后的角色ID，Operation=DELETE 的情况下无意义'"`
	Ctime        int64
	Utime        int64
}

func (u UserRoleLog) TableName() string {
	return "user_role_logs"
}

type UserRoleLogDAO interface {
	BatchCreate(ctx context.Context, userRoleLogs []UserRoleLog) error
}

type userRoleLogDAO struct {
	db *egorm.Component
}

func NewUserRoleLogDAO(db *egorm.Component) UserRoleLogDAO {
	return &userRoleLogDAO{db: db}
}

func (u *userRoleLogDAO) BatchCreate(ctx context.Context, userRoleLogs []UserRoleLog) error {
	if len(userRoleLogs) == 0 {
		return nil
	}
	const batchSize = 100
	now := time.Now().UnixMilli()
	for i := range userRoleLogs {
		userRoleLogs[i].Ctime, userRoleLogs[i].Utime = now, now
	}
	return u.db.WithContext(ctx).CreateInBatches(userRoleLogs, batchSize).Error
}
