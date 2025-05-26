package permission

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/pkg/mqx"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type UserPermissionEventProducer interface {
	Produce(ctx context.Context, event UserPermissionEvent) error
}

func NewUserPermissionEventProducer(producer *kafka.Producer, topic string) (UserPermissionEventProducer, error) {
	return mqx.NewGeneralProducer[UserPermissionEvent](producer, topic)
}

type UserPermissionEvent struct {
	// uid => 全部权限
	Permissions map[int64]UserPermission `json:"permissions"`
}

type UserPermission struct {
	UserID      int64          `json:"userId"`
	BizID       int64          `json:"bizId"`
	Permissions []PermissionV1 `json:"permissions"`
}

type PermissionV1 struct {
	Resource Resource `json:"resource"`
	Action   string   `json:"action"`
	Effect   string   `json:"effect"`
}

type Resource struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}
