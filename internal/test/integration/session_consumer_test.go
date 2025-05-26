//go:build e2e

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"

	"gitee.com/flycash/permission-platform/internal/event/permission"
	"gitee.com/flycash/permission-platform/internal/event/session"
	"gitee.com/flycash/permission-platform/internal/test/ioc"
	confluentkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var testTopic = "test-permission-events"

type SessionConsumerSuite struct {
	suite.Suite
	redisClient redis.Cmdable
	producer    *confluentkafka.Producer
	consumer    *session.Consumer
}

// mqxConsumerWrapper wraps kafka.GoConsumer to implement mqx.Consumer interface

func (s *SessionConsumerSuite) SetupSuite() {
	// Initialize Redis client
	s.redisClient = ioc.InitRedis()

	// Initialize Kafka producer
	s.producer = ioc.InitProducer("test-producer")

	// Initialize Kafka consumer
	config := &confluentkafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "test-group",
		"auto.offset.reset": "earliest",
	}
	kafkaConsumer, err := confluentkafka.NewConsumer(config)
	require.NoError(s.T(), err)
	err = kafkaConsumer.Subscribe(testTopic, nil)
	require.NoError(s.T(), err)
	// Create session consumer
	s.consumer = session.NewConsumer(s.redisClient, kafkaConsumer)
}

func (s *SessionConsumerSuite) TearDownSuite() {
	if s.producer != nil {
		s.producer.Close()
	}
}

func (s *SessionConsumerSuite) TestConsumePermissionEvent() {
	t := s.T()

	event := permission.UserPermissionEvent{
		Permissions: map[int64]permission.UserPermission{
			1: {
				UserID: 1,
				BizID:  1,
				Permissions: []permission.PermissionV1{
					{
						Resource: permission.Resource{
							Key:  "test-resource",
							Type: "test-type",
						},
						Action: "read",
						Effect: "allow",
					},
				},
			},
			2: {
				UserID: 2,
				BizID:  1,
				Permissions: []permission.PermissionV1{
					{
						Resource: permission.Resource{
							Key:  "test-resource",
							Type: "test-type",
						},
						Action: "write",
						Effect: "allow",
					},
				},
			},
		},
	}

	eventBytes, err := json.Marshal(event)
	require.NoError(t, err)

	// Produce message to Kafka
	err = s.producer.Produce(&confluentkafka.Message{
		TopicPartition: confluentkafka.TopicPartition{
			Topic:     &testTopic,
			Partition: confluentkafka.PartitionAny,
		},
		Value: eventBytes,
	}, nil)
	require.NoError(t, err)

	// Flush producer to ensure message is sent
	s.producer.Flush(15 * 1000)

	// Start consumer
	ctx := context.Background()
	go s.consumer.Start(ctx)

	// Wait for consumer to process message
	time.Sleep(5 * time.Second)

	// Verify Redis keys and values
	for uid, expectedPerms := range event.Permissions {
		key := fmt.Sprintf("permission:session:%d", uid)
		val, err := s.redisClient.Get(ctx, key).Result()
		require.NoError(t, err)

		var actualPerms []domain.UserPermission
		err = json.Unmarshal([]byte(val), &actualPerms)
		require.NoError(t, err)
		for idx := range expectedPerms.Permissions {
			require.Equal(t, s.newExpectedPermission(expectedPerms.Permissions[idx], expectedPerms), actualPerms[idx])
		}

	}
}

func TestSessionConsumerSuite(t *testing.T) {
	suite.Run(t, new(SessionConsumerSuite))
}

func (s *SessionConsumerSuite) newExpectedPermission(perm permission.PermissionV1, userPermission permission.UserPermission) domain.UserPermission {
	return domain.UserPermission{
		BizID:  userPermission.BizID,
		UserID: userPermission.UserID,
		Permission: domain.Permission{
			Resource: domain.Resource{
				Type: perm.Resource.Type,
				Key:  perm.Resource.Key,
			},
			Action: perm.Action,
		},
		Effect: domain.Effect(perm.Effect),
	}
}
