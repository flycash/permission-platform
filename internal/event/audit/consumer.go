package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"gitee.com/flycash/permission-platform/internal/pkg/mqx"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	auditdao "gitee.com/flycash/permission-platform/internal/repository/dao/audit"
	"gitee.com/flycash/permission-platform/pkg/canalx"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

type UserRoleBinlogEvent struct {
	Operation    string `json:"operation"`    // 操作类型：INSERT/DELETE
	BizID        int64  `json:"bizId"`        // 业务ID
	UserID       int64  `json:"userId"`       // 用户ID
	BeforeRoleID int64  `json:"beforeRoleId"` // 变更前角色ID，Operation=DELETE 时需要填写
	AfterRoleID  int64  `json:"afterRoleId"`  // 变更后角色ID，Operation=INSERT 时需要填写
}

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

	var evt canalx.Message[dao.UserRole]
	err = json.Unmarshal(msg.Value, &evt)
	if err != nil {
		c.logger.Warn("解析消息失败",
			elog.FieldErr(err),
			elog.Any("msg", msg))
		return err
	}

	if evt.Table != evt.Data[0].TableName() ||
		(evt.Type != "INSERT" && evt.Type != "DELETE") {
		return nil
	}

	err = c.dao.BatchCreate(ctx, slice.Map(evt.Data, func(_ int, src dao.UserRole) auditdao.UserRoleLog {
		var beforeRoleID, afterRoleID int64
		if evt.Type == "INSERT" {
			afterRoleID = src.RoleID
		} else if evt.Type == "DELETE" {
			beforeRoleID = src.RoleID
		}
		return auditdao.UserRoleLog{
			Operation:    evt.Type,
			BizID:        src.BizID,
			UserID:       src.UserID,
			BeforeRoleID: beforeRoleID,
			AfterRoleID:  afterRoleID,
		}
	}))
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
