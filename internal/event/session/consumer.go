package session

import (
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
	"github.com/ecodeclub/ekit/slice"

	"gitee.com/flycash/permission-platform/internal/event/permission"
	"gitee.com/flycash/permission-platform/internal/pkg/mqx"
	"github.com/gotomicro/ego/core/elog"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

const (
	number36       = 36
	defaultTimeout = 5 * time.Second
)

type Consumer struct {
	client   redis.Cmdable
	consumer mqx.Consumer
	logger   *elog.Component
}

func NewConsumer(client redis.Cmdable, consumer mqx.Consumer) *Consumer {
	return &Consumer{
		client:   client,
		consumer: consumer,
		logger:   elog.DefaultLogger,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
			c.Consume(ctx)
			cancel()
		}
	}()
}

func (c *Consumer) Consume(ctx context.Context) {
	msg, err := c.consumer.ReadMessage(-1)
	if err != nil {
		return
	}
	var evt permission.UserPermissionEvent
	err = json.Unmarshal(msg.Value, &evt)
	if err != nil {
		c.logger.Error("解析消息失败",
			elog.FieldErr(err),
			elog.Any("msg", msg))
		return
	}
	vals := make([]any, 0, number36)
	for uid := range evt.Permissions {
		key := c.key(uid)
		userPermission := evt.Permissions[uid]
		domainPermissions := slice.Map(userPermission.Permissions, func(_ int, src permission.PermissionV1) domain.UserPermission {
			return domain.UserPermission{
				BizID:  userPermission.BizID,
				UserID: uid,
				Permission: domain.Permission{
					Resource: domain.Resource{
						Type: src.Resource.Type,
						Key:  src.Resource.Key,
					},
					Action: src.Action,
				},
				Effect: domain.Effect(src.Effect),
			}
		})
		permissionByte, err := json.Marshal(domainPermissions)
		if err != nil {
			c.logger.Error("序列化权限消息失败",
				elog.FieldErr(err),
				elog.Int64("uid", uid),
				elog.Any("permissions", domainPermissions))
			return
		}
		vals = append(vals, key, string(permissionByte))
	}
	err = c.client.MSet(ctx, vals...).Err()
	if err != nil {
		c.logger.Error("设置缓存消息失败",
			elog.FieldErr(err), elog.Any("msg", msg))
		return
	}
	_, err = c.consumer.CommitMessage(msg)
	if err != nil {
		c.logger.Error("提交信息失败",
			elog.FieldErr(err), elog.Any("msg", msg))
		return
	}
}

func (c *Consumer) key(uid int64) string {
	return fmt.Sprintf("permission:session:%d", uid)
}
