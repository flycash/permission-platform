package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"gitee.com/flycash/permission-platform/internal/pkg/mqx"
	auditdao "gitee.com/flycash/permission-platform/internal/repository/dao/audit"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gotomicro/ego/core/elog"
)

type UserRoleBinlogEventConsumer struct {
	consumer mqx.Consumer
	dao      auditdao.UserRoleLogDAO
	logger   *elog.Component
}

func NewUserRoleBinlogEventConsumer(
	consumer *kafka.Consumer,
	dao auditdao.UserRoleLogDAO,
	topic string,
) (*UserRoleBinlogEventConsumer, error) {
	err := consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, err
	}

	return &UserRoleBinlogEventConsumer{
		consumer: consumer,
		dao:      dao,
		logger:   elog.DefaultLogger,
	}, nil
}

func (c *UserRoleBinlogEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			er := c.Consume(ctx)
			if er != nil {
				c.logger.Error("消费用户权限Binlog事件失败", elog.FieldErr(er))
			}
		}
	}()
}

func (c *UserRoleBinlogEventConsumer) Consume(ctx context.Context) error {
	msg, err := c.consumer.ReadMessage(-1)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	var evt UserRoleBinlogEvent
	err = json.Unmarshal(msg.Value, &evt)
	if err != nil {
		c.logger.Warn("解析消息失败",
			elog.FieldErr(err),
			elog.Any("msg", msg))
		return err
	}

	_, err = c.dao.Create(ctx, auditdao.UserRoleLog{
		Operation:    evt.Operation,
		BizID:        evt.BizID,
		UserID:       evt.UserID,
		BeforeRoleID: evt.BeforeRoleID,
		AfterRoleID:  evt.AfterRoleID,
	})
	if err != nil {
		c.logger.Warn("创建用户权限变更操作日志失败",
			elog.FieldErr(err),
			elog.Any("evt", evt))
		return err
	}

	// 消费完成，提交消费进度
	_, err = c.consumer.CommitMessage(msg)
	if err != nil {
		c.logger.Warn("提交消息失败",
			elog.FieldErr(err),
			elog.Any("msg", msg))
		return err
	}
	return nil
}
