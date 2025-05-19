package redis

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/gotomicro/ego/core/elog"
	"github.com/redis/go-redis/v9"
)

type ctxKey string

const (
	bizIDKey    ctxKey = "biz-id"
	uidKey      ctxKey = "uid"
	resourceKey ctxKey = "resource"
)

// RedisOperationType represents the type of Redis operation
type RedisOperationType string

const (
	GET RedisOperationType = "get"
	SET RedisOperationType = "set"
	DEL RedisOperationType = "del"
)

// AccessPlugin is a Redis plugin for permission checking
type AccessPlugin struct {
	client       permissionv1.PermissionServiceClient
	operationMap map[RedisOperationType]string
	logger       *elog.Component
}

type AccessPluginOption func(*AccessPlugin)

func WithOperationMap(operationMap map[RedisOperationType]string) AccessPluginOption {
	return func(p *AccessPlugin) {
		p.operationMap = operationMap
	}
}

// NewAccessPlugin creates a new Redis access plugin
func NewAccessPlugin(client permissionv1.PermissionServiceClient, opts ...AccessPluginOption) *AccessPlugin {
	plugin := &AccessPlugin{
		client: client,
		operationMap: map[RedisOperationType]string{
			GET: "read",
			SET: "write",
			DEL: "delete",
		},
		logger: elog.DefaultLogger,
	}
	for idx := range opts {
		opts[idx](plugin)
	}
	return plugin
}

// DialHook implements redis.Hook interface
func (p *AccessPlugin) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

// ProcessHook implements redis.Hook interface
func (p *AccessPlugin) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// 在执行命令前进行权限校验
		if err := p.checkPermission(ctx, cmd); err != nil {
			return err
		}
		return next(ctx, cmd)
	}
}

// ProcessPipelineHook implements redis.Hook interface
func (p *AccessPlugin) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		// 对管道中的每个命令进行权限校验
		for _, cmd := range cmds {
			if err := p.checkPermission(ctx, cmd); err != nil {
				return err
			}
		}
		return next(ctx, cmds)
	}
}

func (p *AccessPlugin) checkPermission(ctx context.Context, cmd redis.Cmder) error {
	bizID, err := getBizID(ctx)
	if err != nil {
		elog.Warn("get BizID fail", zap.Error(err))
		return nil
	}

	uid, err := getUID(ctx)
	if err != nil {
		elog.Warn("get uid fail", zap.Error(err))
		return nil
	}

	// 根据命令类型获取对应的操作类型
	opType := getOperationType(cmd)
	action, ok := p.operationMap[opType]
	if !ok {
		return nil
	}

	// 获取资源信息
	key := getKeyFromCmd(cmd)
	var resourceType string
	if val, err := getResource(ctx); err == nil {
		key = val.Key
		resourceType = val.Type
	} else {
		resourceType = "redis_key"
	}

	resp, err := p.client.CheckPermission(ctx, &permissionv1.CheckPermissionRequest{
		Uid: uid,
		Permission: &permissionv1.Permission{
			ResourceKey:  key,
			ResourceType: resourceType,
			Actions:      []string{action},
		},
	})
	if err != nil {
		p.logger.Error("权限校验失败",
			elog.FieldErr(err),
			elog.Int64("bizID", bizID),
			elog.Int64("uid", uid),
			elog.String("action", action),
			elog.String("resourceKey", key),
			elog.String("resourceType", resourceType),
		)
		return fmt.Errorf("权限校验失败: %w", err)
	}

	if !resp.Allowed {
		return fmt.Errorf("没有操作权限")
	}

	return nil
}

func getOperationType(cmd redis.Cmder) RedisOperationType {
	switch cmd.Name() {
	case "get":
		return GET
	case "set", "setex", "setnx":
		return SET
	case "del":
		return DEL
	case "exists":
		return GET
	default:
		return ""
	}
}

func getKeyFromCmd(cmd redis.Cmder) string {
	args := cmd.Args()
	if len(args) > 1 {
		if key, ok := args[1].(string); ok {
			return key
		}
	}
	return ""
}

func getBizID(ctx context.Context) (int64, error) {
	value := ctx.Value(bizIDKey)
	if value == nil {
		return 0, fmt.Errorf("biz-id not found in context")
	}

	bizID, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid biz-id type, expected int64 got %T", value)
	}

	return bizID, nil
}

func getUID(ctx context.Context) (int64, error) {
	value := ctx.Value(uidKey)
	if value == nil {
		return 0, fmt.Errorf("uid not found in context")
	}

	uid, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid uid type, expected int64 got %T", value)
	}

	return uid, nil
}

func getResource(ctx context.Context) (*permissionv1.Resource, error) {
	value := ctx.Value(resourceKey)
	if value == nil {
		return nil, fmt.Errorf("resource not found in context")
	}
	res, ok := value.(*permissionv1.Resource)
	if !ok {
		return nil, fmt.Errorf("invalid resource type, expected permissionv1.Resource")
	}
	return res, nil
}
