package audit

import (
	"context"
	"time"

	"github.com/ego-component/egorm"
)

type OperationLog struct {
	ID       int64  `gorm:"primaryKey;autoIncrement;comment:'操作日志表自增ID'"`
	Operator string `gorm:"type:VARCHAR(255);comment:'操作者ID，通常为用户ID'"`
	Key      string `gorm:"type:VARCHAR(255);comment:'业务方内唯一标识，用于标识这次操作请求'"`
	BizID    int64  `gorm:"type:BIGINT;NOT NULL;comment:'表示在该业务ID下调用Method执行Request，Operator的业务ID与此业务ID无关系'"`
	Method   string `gorm:"type:TEXT;NOT NULL;comment:'调用的接口名称'"`
	Request  string `gorm:"type:TEXT;NOT NULL;comment:'请求参数JSON序列化后的字符串'"`
	Ctime    int64
	Utime    int64
}

func (o OperationLog) TableName() string {
	return "operation_logs"
}

type OperationLogDAO interface {
	Create(ctx context.Context, operationLog OperationLog) (int64, error)
}

type operationLogDAO struct {
	db *egorm.Component
}

// NewOperationLogDAO 创建操作日志数据访问对象
func NewOperationLogDAO(db *egorm.Component) OperationLogDAO {
	return &operationLogDAO{
		db: db,
	}
}

func (o *operationLogDAO) Create(ctx context.Context, operationLog OperationLog) (int64, error) {
	now := time.Now().UnixMilli()
	operationLog.Ctime, operationLog.Utime = now, now
	err := o.db.WithContext(ctx).Create(&operationLog).Error
	return operationLog.ID, err
}
