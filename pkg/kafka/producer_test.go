//go:build e2e

package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// TestPermissionServiceClient is a mock implementation of PermissionServiceClient for tests
type TestPermissionServiceClient struct {
	mock.Mock
}

func (m *TestPermissionServiceClient) CheckPermission(ctx context.Context, req *permissionv1.CheckPermissionRequest, _ ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	args := m.Called(ctx, req)
	val, ok := args.Get(0).(*permissionv1.CheckPermissionResponse)
	if !ok {
		return nil, errors.New("CheckPermission fail")
	}
	return val, args.Error(1)
}

func TestAccessProducer_Produce(t *testing.T) {
	// 创建 Kafka 生产者配置
	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "test-producer",
	}

	// 创建 Kafka 生产者
	producer, err := kafka.NewProducer(config)
	require.NoError(t, err)
	defer producer.Close()

	// 创建 GoProducer
	goProducer := NewGoProducer(producer)

	tests := []struct {
		name          string
		setupMocks    func(*TestPermissionServiceClient)
		ctx           context.Context
		msg           *kafka.Message
		expectedError string
	}{
		{
			name: "successful produce",
			setupMocks: func(mpc *TestPermissionServiceClient) {
				mpc.On("CheckPermission", mock.Anything, mock.Anything).Return(&permissionv1.CheckPermissionResponse{
					Allowed: true,
				}, nil)
			},
			ctx: context.WithValue(context.WithValue(context.Background(), bizIDKey, int64(1)), uidKey, int64(100)),
			msg: &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     stringPtr("test-topic"),
					Partition: kafka.PartitionAny,
				},
				Value: []byte("test message"),
			},
			expectedError: "",
		},
		{
			name: "permission denied",
			setupMocks: func(mpc *TestPermissionServiceClient) {
				mpc.On("CheckPermission", mock.Anything, mock.Anything).Return(&permissionv1.CheckPermissionResponse{
					Allowed: false,
				}, nil)
			},
			ctx: context.WithValue(context.WithValue(context.Background(), bizIDKey, int64(1)), uidKey, int64(100)),
			msg: &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     stringPtr("test-topic"),
					Partition: kafka.PartitionAny,
				},
				Value: []byte("test message"),
			},
			expectedError: "权限校验未通过: 用户 100 没有向 topic test-topic 发送消息的权限",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 TestPermissionServiceClient
			testPermissionClient := new(TestPermissionServiceClient)
			tt.setupMocks(testPermissionClient)

			// 创建 AccessProducer
			accessProducer := NewProducer(
				goProducer,
				testPermissionClient,
				func(msg *kafka.Message) (string, string) {
					return "kafka", *msg.TopicPartition.Topic
				},
			)

			// 创建交付通道
			deliveryChan := make(chan kafka.Event)

			// 发送消息
			err = accessProducer.Produce(tt.ctx, tt.msg, deliveryChan)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)

				// 等待消息发送完成
				select {
				case e := <-deliveryChan:
					ev, ok := e.(*kafka.Message)
					require.True(t, ok)

					assert.Equal(t, "test-topic", *ev.TopicPartition.Topic)
				case <-time.After(5 * time.Second):
					t.Fatal("Message delivery timeout")
				}
			}

			// 验证权限检查被调用
			testPermissionClient.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
