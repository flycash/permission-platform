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

func (s *SessionConsumerSuite) SetupSuite() {
	// Initialize Redis client
	s.redisClient = ioc.InitRedis()

	// Initialize Kafka producer
	s.producer = ioc.InitProducer("test-producer")

	// Initialize Kafka consumer
	config := &confluentkafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
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

	// Set up session ID mapping in Redis
	sessionIDMap := map[string]string{
		"1": "session-1",
		"2": "session-2",
	}
	for uid, sessionID := range sessionIDMap {
		err := s.redisClient.Set(context.Background(), uid, sessionID, 0).Err()
		require.NoError(t, err)
	}

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
	time.Sleep(10 * time.Second)

	// Verify Redis keys and values
	for uid, expectedPerms := range event.Permissions {
		sessionID := sessionIDMap[fmt.Sprintf("%d", uid)]
		hashKey := fmt.Sprintf("session:%s", sessionID)

		// Get all fields from the hash
		vals, err := s.redisClient.HGetAll(ctx, hashKey).Result()
		require.NoError(t, err)
		require.NotEmpty(t, vals)

		// Verify the permissions are stored correctly
		for _, val := range vals {
			var actualPerms []domain.UserPermission
			err = json.Unmarshal([]byte(val), &actualPerms)
			require.NoError(t, err)

			// Verify each permission
			for idx := range expectedPerms.Permissions {
				require.Equal(t, s.newExpectedPermission(expectedPerms.Permissions[idx], expectedPerms), actualPerms[idx])
			}
		}
	}
}

func TestSessionConsumerSuite(t *testing.T) {
	t.Skip()
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
