package audit

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/pkg/mqx"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type UserRoleBinlogEvent struct {
	Operation    string `json:"operation"`    // 操作类型：INSERT/DELETE
	BizID        int64  `json:"bizId"`        // 业务ID
	UserID       int64  `json:"userId"`       // 用户ID
	BeforeRoleID int64  `json:"beforeRoleId"` // 变更前角色ID，Operation=DELETE 时需要填写
	AfterRoleID  int64  `json:"afterRoleId"`  // 变更后角色ID，Operation=INSERT 时需要填写
}

//go:generate mockgen -source=./producer.go -package=evtmocks -destination=../mocks/user_role_event_producer.mock.go -typed UserRoleBinlogEventProducer
type UserRoleBinlogEventProducer interface {
	Produce(ctx context.Context, evt UserRoleBinlogEvent) error
}

func NewUserRoleBinlogEventProducer(producer *kafka.Producer, topic string) (UserRoleBinlogEventProducer, error) {
	return mqx.NewGeneralProducer[UserRoleBinlogEvent](producer, topic)
}
