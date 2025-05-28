package kafka

import (
	"context"
	"fmt"
	"gitee.com/flycash/permission-platform/pkg/ctxx"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gotomicro/ego/core/elog"
)


const (
	produce  string = "produce"
)

type AccessProducer struct {
	producer     Producer
	client       permissionv1.PermissionServiceClient
	logger       *elog.Component
	actions      map[string]string
	resourceFunc func(msg *kafka.Message) (string, string) // 第一个是resourceType，第二个resourceValue
}
type AccessProducerOption func(*AccessProducer)

func WithActions(actions map[string]string) AccessProducerOption {
	return func(producer *AccessProducer) {
		producer.actions = actions
	}
}

func NewProducer(p Producer,
	client permissionv1.PermissionServiceClient,
	resourceFunc func(msg *kafka.Message) (string, string),
	opts ...AccessProducerOption,
) *AccessProducer {
	accessProducer := &AccessProducer{
		producer:     p,
		client:       client,
		logger:       elog.DefaultLogger,
		resourceFunc: resourceFunc,
		actions: map[string]string{
			produce: "produce",
		},
	}
	for idx := range opts {
		opts[idx](accessProducer)
	}
	return accessProducer
}

func (p *AccessProducer) Produce(ctx context.Context, msg *kafka.Message, deliveryChan chan kafka.Event) error {
	// 从context中获取bizID和uid
	bizID, err := ctxx.GetBizID(ctx)
	if err != nil {
		p.logger.Error("获取bizID失败",
			elog.FieldErr(err),
			elog.String("topic", *msg.TopicPartition.Topic),
		)
		return err
	}

	uid, err :=ctxx.GetUID(ctx)
	if err != nil {
		p.logger.Error("获取uid失败",
			elog.FieldErr(err),
			elog.String("topic", *msg.TopicPartition.Topic),
		)
		return err
	}
	typ, key := p.resourceFunc(msg)
	// 构建权限检查请求
	req := &permissionv1.CheckPermissionRequest{
		Uid: uid,
		Permission: &permissionv1.Permission{
			ResourceType: typ,
			ResourceKey:  key,
			Actions:      []string{p.action()},
		},
	}

	// 检查权限
	resp, err := p.client.CheckPermission(ctx, req)
	if err != nil {
		p.logger.Error("权限校验失败",
			elog.FieldErr(err),
			elog.Int64("bizID", bizID),
			elog.Int64("uid", uid),
			elog.String("action", produce),
			elog.String("resourceKey", *msg.TopicPartition.Topic),
			elog.String("resourceType", "kafka"),
		)
		return fmt.Errorf("权限校验失败: %w", err)
	}

	if !resp.Allowed {
		p.logger.Error("权限校验未通过",
			elog.Int64("bizID", bizID),
			elog.Int64("uid", uid),
			elog.String("action", produce),
			elog.String("resourceKey", *msg.TopicPartition.Topic),
			elog.String("resourceType", "kafka"),
		)
		return fmt.Errorf("权限校验未通过: 用户 %d 没有向 topic %s 发送消息的权限", uid, *msg.TopicPartition.Topic)
	}

	return p.producer.Produce(ctx, msg, deliveryChan)
}

func (p *AccessProducer) action() string {
	return p.actions[produce]
}

