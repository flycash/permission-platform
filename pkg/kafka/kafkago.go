package kafka

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type GoProducer struct {
	*kafka.Producer
}

func (k *GoProducer) Produce(_ context.Context, msg *kafka.Message, deliveryChan chan kafka.Event) error {
	return k.Producer.Produce(msg, deliveryChan)
}

func NewGoProducer(producer *kafka.Producer) *GoProducer {
	return &GoProducer{producer}
}

type GoConsumer struct {
	*kafka.Consumer
}

func NewGoConsumer(c *kafka.Consumer) *GoConsumer {
	return &GoConsumer{c}
}

func (k *GoConsumer) Subscribe(_ context.Context, topic string, rebalanceCb kafka.RebalanceCb) error {
	return k.Consumer.Subscribe(topic, rebalanceCb)
}

func (k *GoConsumer) ReadMessage(_ context.Context, timeout time.Duration) (*kafka.Message, error) {
	return k.Consumer.ReadMessage(timeout)
}
