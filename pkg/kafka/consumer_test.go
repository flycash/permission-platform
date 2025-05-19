//go:build e2e

package kafka

import (
	"context"
	"strings"
	"testing"
	"time"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// newTestContext creates a new context with bizID and uid
func newTestContext(bizID, uid int64) context.Context {
	ctx := context.Background()
	if bizID != 0 {
		ctx = context.WithValue(ctx, bizIDKey, bizID)
	}
	if uid != 0 {
		ctx = context.WithValue(ctx, uidKey, uid)
	}
	return ctx
}

// MockPermissionServiceClient 是一个简单的权限服务客户端模拟实现
type MockPermissionServiceClient struct{}

func (m *MockPermissionServiceClient) CheckPermission(ctx context.Context, req *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	// 如果 resourceKey 包含 "pass"，则允许访问
	if strings.Contains(req.Permission.ResourceKey, "deny") {
		return &permissionv1.CheckPermissionResponse{
			Allowed: false,
		}, nil
	}
	return &permissionv1.CheckPermissionResponse{
		Allowed: true,
	}, nil
}

func TestAccessConsumer_Subscribe(t *testing.T) {
	t.Skip("Skipping integration test")

	// 创建 Kafka 消费者配置
	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9094",
		"group.id":          "test-group",
		"auto.offset.reset": "earliest",
	}

	// 创建 Kafka 消费者
	consumer, err := kafka.NewConsumer(config)
	require.NoError(t, err)
	defer consumer.Close()

	// 创建 GoConsumer
	goConsumer := NewGoConsumer(consumer)

	tests := []struct {
		name          string
		setupMocks    func(*TestPermissionServiceClient)
		ctx           context.Context
		topic         string
		expectedError string
	}{
		{
			name: "successful subscription",
			setupMocks: func(tpc *TestPermissionServiceClient) {
				tpc.On("CheckPermission", mock.Anything, mock.Anything).Return(&permissionv1.CheckPermissionResponse{
					Allowed: true,
				}, nil)
			},
			ctx:           newTestContext(1, 100),
			topic:         "test-topic-",
			expectedError: "",
		},
		{
			name: "permission denied",
			setupMocks: func(tpc *TestPermissionServiceClient) {
				tpc.On("CheckPermission", mock.Anything, mock.Anything).Return(&permissionv1.CheckPermissionResponse{
					Allowed: false,
				}, nil)
			},
			ctx:           newTestContext(1, 100),
			topic:         "test-topic",
			expectedError: "权限校验未通过: 用户 100 没有订阅 topic test-topic 的权限",
		},
		{
			name: "missing bizID",
			setupMocks: func(tpc *TestPermissionServiceClient) {
				// No mock setup needed
			},
			ctx:           newTestContext(0, 100),
			topic:         "test-topic",
			expectedError: "权限校验未通过: 未找到bizID",
		},
		{
			name: "missing uid",
			setupMocks: func(tpc *TestPermissionServiceClient) {
				// No mock setup needed
			},
			ctx:           newTestContext(1, 0),
			topic:         "test-topic",
			expectedError: "权限校验未通过: 未找到uid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 TestPermissionServiceClient
			testPermissionClient := new(TestPermissionServiceClient)
			tt.setupMocks(testPermissionClient)

			// 创建 AccessConsumer
			accessConsumer := NewAccessConsumer(
				goConsumer,
				testPermissionClient,
				func(msg *kafka.Message) (string, string) {
					return "kafka", *msg.TopicPartition.Topic
				},
			)

			// 测试订阅
			err := accessConsumer.Subscribe(tt.ctx, tt.topic, nil)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			// 验证权限检查被调用
			testPermissionClient.AssertExpectations(t)
		})
	}
}

// initKafkaTopic 初始化 Kafka topic
func initKafkaTopic(t *testing.T, bootstrapServers string, topic string) {
	// 创建 Kafka 管理员配置
	adminConfig := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	}

	// 创建 Kafka 管理员客户端
	admin, err := kafka.NewAdminClient(adminConfig)
	require.NoError(t, err)
	defer admin.Close()

	// 创建 topic
	results, err := admin.CreateTopics(
		context.Background(),
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}},
	)
	require.NoError(t, err)

	// 检查 topic 创建结果
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			t.Fatalf("Failed to create topic %s: %v", result.Topic, result.Error)
		}
	}
}

func TestAccessConsumer_ReadMessage(t *testing.T) {
	t.Skip("还在修改")
	bootstrapServers := "localhost:9092"
	testTopic := "test-read"

	// 初始化 topic
	initKafkaTopic(t, bootstrapServers, testTopic)

	// 创建 Kafka 消费者配置
	config := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          "test-group",
		"auto.offset.reset": "earliest",
	}

	// 创建 Kafka 消费者
	consumer, err := kafka.NewConsumer(config)
	require.NoError(t, err)
	defer consumer.Close()

	// 创建 Kafka 生产者配置
	producerConfig := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "test-producer",
	}

	// 创建 Kafka 生产者
	producer, err := kafka.NewProducer(producerConfig)
	require.NoError(t, err)
	defer producer.Close()

	// 创建 GoConsumer
	goConsumer := NewGoConsumer(consumer)

	// 测试消息
	testMessages := []struct {
		key   string
		value string
	}{
		{
			key:   "key1",
			value: "message1",
		},
		{
			key:   "key2-deny",
			value: "message2",
		},
		{
			key:   "key3",
			value: "message3",
		},
	}

	// 发送测试消息
	for _, msg := range testMessages {
		topic := testTopic
		err = producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(msg.key),
			Value: []byte(msg.value),
		}, nil)
		require.NoError(t, err)
	}

	// 等待所有消息发送完成
	producer.Flush(15 * 1000)

	// 创建 MockPermissionServiceClient
	mockClient := &MockPermissionServiceClient{}

	// 创建 AccessConsumer
	accessConsumer := NewAccessConsumer(
		goConsumer,
		mockClient,
		func(msg *kafka.Message) (string, string) {
			return "kafka", *msg.TopicPartition.Topic + ":" + string(msg.Key)
		},
	)

	time.Sleep(10 * time.Second)
	// 订阅测试主题
	err = accessConsumer.Subscribe(newTestContext(1, 100), testTopic, nil)
	require.NoError(t, err)

	// 读取第一条消息（应该有权限，因为 key 不包含 "deny"）
	msg1, err := accessConsumer.ReadMessage(newTestContext(1, 100), 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, testTopic, *msg1.TopicPartition.Topic)
	assert.Equal(t, "key1", string(msg1.Key))
	assert.Equal(t, "message1", string(msg1.Value))

	// 读取第二条消息（应该被跳过，因为 key 包含 "deny"）
	// 读取第三条消息（应该有权限，因为 key 不包含 "deny"）
	msg2, err := accessConsumer.ReadMessage(newTestContext(1, 100), 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, testTopic, *msg2.TopicPartition.Topic)
	assert.Equal(t, "key3", string(msg2.Key))
	assert.Equal(t, "message3", string(msg2.Value))
}
