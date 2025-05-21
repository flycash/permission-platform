package failover

import (
	"context"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/pkg/database/monitor"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gotomicro/ego/core/elog"
)

type baseConsumer struct {
	consumer           *kafka.Consumer
	dbMonitor          monitor.DBMonitor
	logger             *elog.Component
	topic              string
	processMessageFunc func(ctx context.Context, msg *kafka.Message) error
}

func newBaseConsumer(consumer *kafka.Consumer, dbMonitor monitor.DBMonitor, processMessageFunc func(ctx context.Context, msg *kafka.Message) error) *baseConsumer {
	return &baseConsumer{
		consumer:           consumer,
		dbMonitor:          dbMonitor,
		logger:             elog.DefaultLogger,
		topic:              FailoverTopic,
		processMessageFunc: processMessageFunc,
	}
}

// Start 在后台协程中开始消费Kafka消息
func (b *baseConsumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				b.logger.Info("ConnPoolEventConsumer因上下文取消而停止")
				return
			default:
				err := b.Consume(ctx)
				if err != nil {
					b.logger.Error("消费消息失败", elog.FieldErr(err))
				}
			}
		}
	}()
	b.logger.Info("正在启动ConnPoolEventConsumer，监听主题", elog.String("topic", FailoverTopic))
}

// Consume 处理单个消息
func (b *baseConsumer) Consume(ctx context.Context) error {
	// 检查数据库健康状态
	if !b.dbMonitor.Health() {
		// 获取当前消费者分配的分区
		assigned, err := b.consumer.Assignment()
		if err != nil {
			return fmt.Errorf("获取消费者已分配分区失败: %w", err)
		}
		// 如果有分配的分区，暂停它们
		if len(assigned) > 0 {
			if err := b.consumer.Pause(assigned); err != nil {
				return fmt.Errorf("暂停分区失败: %w", err)
			}
			b.logger.Info("已暂停消费者分配的分区", elog.Int("分区数量", len(assigned)))

			// 等待2秒
			time.Sleep(defaultSleepTime)

			// 恢复分区，即使数据库仍然不健康也恢复分区
			// 因为下一次消费循环会再次检查并暂停
			if err := b.consumer.Resume(assigned); err != nil {
				return fmt.Errorf("恢复分区失败: %w", err)
			}
			b.logger.Info("已恢复消费者分配的分区", elog.Int("分区数量", len(assigned)))
		}

		return nil
	}

	// 数据库健康，获取并处理消息
	ev := b.consumer.Poll(poll)
	if ev == nil {
		return nil // 没有可用消息，稍后重试
	}

	switch e := ev.(type) {
	case *kafka.Message:
		if err := b.processMessageFunc(ctx, e); err != nil {
			b.logger.Error("处理消息失败",
				elog.FieldErr(err),
				elog.String("主题", *e.TopicPartition.Topic),
				elog.String("分区", fmt.Sprintf("%d", e.TopicPartition.Partition)),
				elog.String("偏移量", fmt.Sprintf("%d", e.TopicPartition.Offset)))
			return err
		}

		// 提交消息
		if _, err := b.consumer.CommitMessage(e); err != nil {
			return fmt.Errorf("提交消息失败: %w", err)
		}

	case kafka.Error:
		return fmt.Errorf("kafka错误: %w", e)
	}

	return nil
}

// Stop 停止消费者
func (b *baseConsumer) Stop() error {
	return b.consumer.Close()
}
