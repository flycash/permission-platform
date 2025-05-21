package failover

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/pkg/database/monitor"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gotomicro/ego/core/elog"
	"gorm.io/gorm"
)

const (
	defaultSleepTime = 2 * time.Second
	poll             = 1000
)

type ConnPoolEventConsumer struct {
	consumer *baseConsumer
	db       *gorm.DB
	logger   *elog.Component
}

// NewConnPoolEventConsumer 创建一个新的连接池事件消费者
func NewConnPoolEventConsumer(consumer *kafka.Consumer, db *gorm.DB, dbMonitor monitor.DBMonitor) *ConnPoolEventConsumer {
	c := &ConnPoolEventConsumer{
		db:     db,
		logger: elog.DefaultLogger,
	}
	c.consumer = newBaseConsumer(consumer, dbMonitor, c.processMessage)
	return c
}

// Start 在后台协程中开始消费Kafka消息
func (c *ConnPoolEventConsumer) Start(ctx context.Context) {
	c.consumer.Start(ctx)
}

// processMessage 处理单个ConnPoolEvent消息
func (c *ConnPoolEventConsumer) processMessage(ctx context.Context, msg *kafka.Message) error {
	// 反序列化消息
	var event ConnPoolEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("反序列化消息失败: %w", err)
	}

	c.logger.Info("正在处理ConnPoolEvent",
		elog.String("sql", event.SQL),
		elog.Any("参数", event.Args))

	// 在数据库上执行SQL
	_, err := c.db.ConnPool.ExecContext(ctx, event.SQL, event.Args...)
	if err != nil {
		return fmt.Errorf("执行事件中的SQL失败: %w", err)
	}

	return nil
}
