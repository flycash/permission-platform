package audit

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gotomicro/ego/core/elog"
	"github.com/withlin/canal-go/client"
	"github.com/withlin/canal-go/protocol/entry"
	"google.golang.org/protobuf/proto"
)

const (
	InsertOperation = "INSERT"
	DeleteOperation = "DELETE"
)

type CanalUserRoleBinlogEventProducer struct {
	conn            client.CanalConnector
	producer        UserRoleBinlogEventProducer
	minLoopDuration time.Duration
	batchSize       int32
	timeOut         int64
	units           int32

	logger *elog.Component
}

// NewCanalUserRoleBinlogEventProducer 创建通过Canal监听用户权限表的增删事件转发到Kafka任务
func NewCanalUserRoleBinlogEventProducer(
	conn client.CanalConnector,
	producer UserRoleBinlogEventProducer,
	minLoopDuration time.Duration,
	batchSize int32,
	timeOut int64,
	units int32,
) *CanalUserRoleBinlogEventProducer {
	return &CanalUserRoleBinlogEventProducer{
		conn:            conn,
		producer:        producer,
		minLoopDuration: minLoopDuration,
		batchSize:       batchSize,
		timeOut:         timeOut,
		units:           units,
		logger:          elog.DefaultLogger.With(elog.FieldName("user_role_binlog_task")),
	}
}

func (t *CanalUserRoleBinlogEventProducer) Start(ctx context.Context) {
	now := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if d := time.Since(now); d < t.minLoopDuration {
			time.Sleep(t.minLoopDuration - d)
		}

		now = time.Now()
		message, err := t.conn.Get(t.batchSize, &t.timeOut, &t.units)
		if err != nil {
			t.logger.Error("获取Canal消息失败",
				elog.FieldErr(err),
			)
			continue
		}

		if len(message.Entries) == 0 {
			continue
		}

		err = t.handleEntries(ctx, message.Entries)
		if err != nil {
			t.logger.Error("消费Canal消息失败",
				elog.FieldErr(err),
				elog.Any("message", message),
			)
			continue
		}

		err = t.conn.Ack(message.Id)
		if err != nil {
			t.logger.Error("确认Canal消息失败",
				elog.FieldErr(err),
				elog.Any("message", message),
			)
		}
	}
}

func (t *CanalUserRoleBinlogEventProducer) handleEntries(ctx context.Context, entries []entry.Entry) error {
	for i := range entries {

		if entries[i].GetEntryType() != entry.EntryType_ROWDATA {
			continue
		}

		change := &entry.RowChange{}
		if err := proto.Unmarshal(entries[i].GetStoreValue(), change); err != nil {
			return fmt.Errorf("反序列化行变化失败: %w", err)
		}

		if change.GetEventType() != entry.EventType_INSERT &&
			change.GetEventType() != entry.EventType_DELETE {
			continue
		}

		// 记录事件信息
		log.Printf("[binlog-syncer] 收到 %s 事件", change.GetDdlSchemaName())

		for _, row := range change.GetRowDatas() {
			var evt UserRoleBinlogEvent
			//nolint:exhaustive // 忽略
			switch change.GetEventType() {
			case entry.EventType_INSERT:
				evt = t.newEvent(InsertOperation, row.GetAfterColumns())
			case entry.EventType_DELETE:
				evt = t.newEvent(DeleteOperation, row.GetBeforeColumns())
			}
			if err := t.producer.Produce(ctx, evt); err != nil {
				return fmt.Errorf("向kafka发送消息失败: %w", err)
			}
		}
	}
	return nil
}

func (t *CanalUserRoleBinlogEventProducer) newEvent(operation string, cols []*entry.Column) UserRoleBinlogEvent {
	mp := make(map[string]string, len(cols))
	for _, c := range cols {
		mp[c.Name] = c.Value
	}

	toInt64 := func(s string) int64 {
		v, _ := strconv.ParseInt(s, 10, 64)
		return v
	}
	var beforeRoleID, afterRoleID int64

	switch operation {
	case InsertOperation:
		afterRoleID = toInt64(mp["role_id"])
	case DeleteOperation:
		beforeRoleID = toInt64(mp["role_id"])
	}

	return UserRoleBinlogEvent{
		Operation:    operation,
		BizID:        toInt64(mp["biz_id"]),
		UserID:       toInt64(mp["user_id"]),
		BeforeRoleID: beforeRoleID,
		AfterRoleID:  afterRoleID,
	}
}
