package kafka

import (
	"context"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/pkg/ctxx"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gotomicro/ego/core/elog"
)

const (
	consume   string = "consume"
	subscribe string = "subscribe"
)

type AccessConsumer struct {
	consumer     Consumer
	client       permissionv1.PermissionServiceClient
	logger       *elog.Component
	resourceFunc func(msg *kafka.Message) (string, string) // 第一个是resourceType，第二个resourceValue
	actions      map[string]string
}

func (c *AccessConsumer) CommitMessage(m *kafka.Message) ([]kafka.TopicPartition, error) {
	return c.consumer.CommitMessage(m)
}

func NewAccessConsumer(c Consumer, client permissionv1.PermissionServiceClient,
	resourceFunc func(msg *kafka.Message) (string, string),
) *AccessConsumer {
	return &AccessConsumer{
		consumer: c,
		client:   client,
		logger:   elog.DefaultLogger,
		actions: map[string]string{
			consume: consume,
		},
		resourceFunc: resourceFunc,
	}
}

// ReadMessage 读取消息，带权限控制
func (c *AccessConsumer) ReadMessage(ctx context.Context, timeout time.Duration) (*kafka.Message, error) {
	for {
		startTime := time.Now().UnixMilli()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		msg, err := c.consumer.ReadMessage(ctx, timeout)
		if err != nil {
			return nil, err
		}

		bizID, err := ctxx.GetBizID(ctx)
		if err != nil {
			c.logger.Error("获取bizID失败",
				elog.FieldErr(err),
				elog.String("topic", *msg.TopicPartition.Topic),
			)
			return nil, err
		}

		uid, err := ctxx.GetUID(ctx)
		if err != nil {
			c.logger.Error("获取uid失败",
				elog.FieldErr(err),
				elog.String("topic", *msg.TopicPartition.Topic),
			)
			return nil, err
		}

		typ, key := c.resourceFunc(msg)
		req := &permissionv1.CheckPermissionRequest{
			Uid: uid,
			Permission: &permissionv1.Permission{
				ResourceType: typ,
				ResourceKey:  key,
				Actions:      []string{c.action()},
			},
		}

		resp, err := c.client.CheckPermission(ctx, req)
		if err != nil {
			c.logger.Error("权限校验失败",
				elog.FieldErr(err),
				elog.Int64("bizID", bizID),
				elog.Int64("uid", uid),
				elog.String("action", consume),
				elog.String("resourceKey", key),
				elog.String("resourceType", typ),
			)
			return nil, fmt.Errorf("权限校验失败: %w", err)
		}

		if !resp.Allowed {
			// 判断有没有到超时时间
			if timeout > 0 {
				endTime := time.Now().UnixMilli()
				spendTime := time.Duration(endTime-startTime) * time.Millisecond
				// 修改超时时间
				if spendTime < timeout {
					timeout = timeout - spendTime
				} else {
					return nil, context.DeadlineExceeded
				}
			}
			// 忽略此消息
			_, err = c.consumer.CommitMessage(msg)
			if err != nil {
				c.logger.Error("提交消息失败",
					elog.FieldErr(err),
					elog.Int64("bizID", bizID),
					elog.Int64("uid", uid),
					elog.String("action", consume),
					elog.String("resourceKey", key),
					elog.String("resourceType", typ),
					elog.String("message", string(msg.Value)),
				)
				return nil, fmt.Errorf("权限校验失败: %w", err)
			}
			continue
		}
		return msg, nil
	}
}

// Subscribe 订阅主题，带权限控制
func (c *AccessConsumer) Subscribe(ctx context.Context, topic string, rebalanceCb kafka.RebalanceCb) error {
	// 从context中获取bizID和uid
	bizID, err := ctxx.GetBizID(ctx)
	if err != nil {
		return fmt.Errorf("权限校验未通过: 未找到bizID")
	}
	uid, err := ctxx.GetUID(ctx)
	if err != nil {
		return fmt.Errorf("权限校验未通过: 未找到uid")
	}
	// 构建权限检查请求
	req := &permissionv1.CheckPermissionRequest{
		Uid: uid,
		Permission: &permissionv1.Permission{
			ResourceType: "kafka",
			ResourceKey:  topic,
			Actions:      []string{subscribe},
		},
	}

	// 检查权限
	resp, err := c.client.CheckPermission(ctx, req)
	if err != nil {
		c.logger.Error("权限校验失败",
			elog.FieldErr(err),
			elog.Int64("bizID", bizID),
			elog.Int64("uid", uid),
			elog.String("action", consume),
			elog.String("resourceKey", topic),
			elog.String("resourceType", "kafka"),
		)
		return fmt.Errorf("权限校验失败: %w", err)
	}

	if !resp.Allowed {
		c.logger.Error("权限校验未通过",
			elog.Int64("bizID", bizID),
			elog.Int64("uid", uid),
			elog.String("action", consume),
			elog.String("resourceKey", topic),
			elog.String("resourceType", "kafka"),
		)
		return fmt.Errorf("权限校验未通过: 用户 %d 没有订阅 topic %s 的权限", uid, topic)
	}

	return c.consumer.Subscribe(ctx, topic, rebalanceCb)
}

func (c *AccessConsumer) action() string {
	return c.actions[consume]
}
