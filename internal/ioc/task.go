package ioc

import (
	"gitee.com/flycash/permission-platform/internal/event/audit"
)

func InitTasks(t1 *audit.CanalUserRoleBinlogEventProducer,
	t2 *audit.UserRoleBinlogEventConsumer,
) []Task {
	return []Task{
		t1,
		t2,
	}
}
