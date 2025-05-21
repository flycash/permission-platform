package failover

import (
	"context"

	"gitee.com/flycash/permission-platform/internal/pkg/database/monitor"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gotomicro/ego/core/elog"
	"github.com/redis/go-redis/v9"
)

type ConsumerRedis struct {
	consumer     *baseConsumer
	client       redis.Cmdable
	logger       *elog.Component
	resourceFunc func(msg *kafka.Message) (key string, val []byte, ok bool)
}

func (c *ConsumerRedis) Start(ctx context.Context) {
	c.consumer.Start(ctx)
}

// NewConsumerRedis 创建一个新的连接池事件消费者
func NewConsumerRedis(consumer *kafka.Consumer, client redis.Cmdable, dbMonitor monitor.DBMonitor, resourceFunc func(msg *kafka.Message) (key string, val []byte, ok bool)) *ConsumerRedis {
	c := &ConsumerRedis{
		logger:       elog.DefaultLogger,
		client:       client,
		resourceFunc: resourceFunc,
	}
	c.consumer = newBaseConsumer(consumer, dbMonitor, c.processMessage)
	return c
}

// processMessage 处理单个ConnPoolEvent消息
func (c *ConsumerRedis) processMessage(ctx context.Context, msg *kafka.Message) error {
	key, val, ok := c.resourceFunc(msg)
	if !ok {
		return nil
	}
	return c.client.Set(ctx, key, val, -1).Err()
}
