package ioc

import (
	"context"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/event/failover"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	maxInterval = 10 * time.Second
	number1     = 1
)

func InitTopic() {
	topics := []kafka.TopicSpecification{
		{
			Topic:         failover.FailoverTopic,
			NumPartitions: number1,
		},
	}
	initTopic(topics...)
}

func InitProducer(id string) *kafka.Producer {
	// 初始化生产者
	config := &kafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
		"client.id":         id,
	}

	// 2. 创建生产者实例
	producer, err := kafka.NewProducer(config)
	if err != nil {
		panic(fmt.Sprintf("创建生产者失败: %v", err))
	}
	return producer
}

func initTopic(topics ...kafka.TopicSpecification) {
	// 创建 AdminClient
	const kafkaAddr = "127.0.0.1:9092"
	const serverName = "bootstrap.servers"
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		serverName: kafkaAddr,
	})
	if err != nil {
		panic(fmt.Sprintf("创建kafka连接失败: %v", err))
	}
	defer adminClient.Close()
	// 设置要创建的主题的配置信息
	ctx, cancel := context.WithTimeout(context.Background(), maxInterval)
	defer cancel()
	// 创建主题
	results, err := adminClient.CreateTopics(
		ctx,
		topics,
	)
	if err != nil {
		panic(fmt.Sprintf("创建topic失败: %v", err))
	}

	// 处理创建主题的结果
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("创建topic失败 %s: %v\n", result.Topic, result.Error)
		} else {
			fmt.Printf("topic %s 创建成功\n", result.Topic)
		}
	}
}
