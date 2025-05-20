package kafka

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer interface {
	Produce(ctx context.Context, msg *kafka.Message, deliveryChan chan kafka.Event) error
}

type Consumer interface {
	// 订阅
	Subscribe(ctx context.Context, topic string, rebalanceCb kafka.RebalanceCb) error
	ReadMessage(ctx context.Context, timeout time.Duration) (*kafka.Message, error)
	CommitMessage(m *kafka.Message) ([]kafka.TopicPartition, error)
}
