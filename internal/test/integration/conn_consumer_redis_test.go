//go:build e2e

package integration

import (
	"fmt"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/internal/event/failover"
	monMocks "gitee.com/flycash/permission-platform/internal/pkg/database/monitor/mocks"
	"gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type User0 struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"column:name"`
}

type ConsumerRedisSuite struct {
	suite.Suite
	mockCtrl    *gomock.Controller
	mockMonitor *monMocks.MockDBMonitor
	client      redis.Cmdable
	producer    failover.ConnPoolEventProducer
	consumer    *failover.ConsumerRedis
}

func (s *ConsumerRedisSuite) SetupSuite() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockMonitor = monMocks.NewMockDBMonitor(s.mockCtrl)

	// 初始化生产者
	ioc.InitTopic()
	pro := ioc.InitProducer("failover")
	s.producer = failover.NewProducer(pro)

	redisClient := ioc.InitRedis()
	s.client = redisClient

	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "failOverGroup",
		"auto.offset.reset": "earliest",
	}
	con, err := kafka.NewConsumer(config)
	require.NoError(s.T(), err)
	err = con.SubscribeTopics([]string{failover.FailoverTopic}, nil)
	require.NoError(s.T(), err)
	// 触发分区
	s.initializeConsumer(con)
	// 创建消费者实例
	redisFunc := func(msg *kafka.Message) (key string, val []byte, ok bool) {
		var event failover.ConnPoolEvent
		err = json.Unmarshal(msg.Value, &event)
		require.NoError(s.T(), err)
		keyAny := event.Args[1].(string)
		valAny := event.Args[1].(string)
		return keyAny, []byte(valAny), true
	}
	s.consumer = failover.NewConsumerRedis(con, redisClient, s.mockMonitor, redisFunc)
}

func (c *ConsumerRedisSuite) initializeConsumer(con *kafka.Consumer) {
	// 设置超时时间，避免永久阻塞
	timeout := time.Now().Add(30 * time.Second)
	for time.Now().Before(timeout) {
		// 进行一次Poll操作，触发分区分配
		ev := con.Poll(1000)
		if ev == nil {
			continue
		}
		return
	}
}

func (s *ConsumerRedisSuite) TestConsumerBehaviorWithMonitor() {
	t := s.T()
	start := time.Now().Unix()
	// 第一阶段：监控状态为false时的消费行为
	s.mockMonitor.EXPECT().Health().DoAndReturn(func() bool {
		end := time.Now().Unix()
		if end-start > 3 {
			return true
		}
		return false
	}).AnyTimes()

	// 推送三条测试消息
	events := []failover.ConnPoolEvent{
		{SQL: "INSERT INTO user0 (`id`,`name`) VALUES (?,?)", Args: []any{1, "user1"}},
		{SQL: "INSERT INTO user0 (`id`,`name`) VALUES (?,?)", Args: []any{2, "user2"}},
		{SQL: "INSERT INTO user0 (`id`,`name`) VALUES (?,?)", Args: []any{3, "user3"}},
	}
	for _, event := range events {
		err := s.producer.Produce(t.Context(), event)
		require.NoError(t, err)
	}

	// 消费者启动
	go s.consumer.Start(t.Context())
	time.Sleep(6 * time.Second)

	// 验证数据正确写入
	for i := 1; i <= 3; i++ {
		err := s.client.Get(t.Context(), fmt.Sprintf("user%d", i)).Err()
		require.NoError(t, err)
	}
}

func TestConsumerRedisSuite(t *testing.T) {
	suite.Run(t, new(ConsumerRedisSuite))
}
