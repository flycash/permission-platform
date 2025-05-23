package gorm

import (
	"context"
	"fmt"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/gotomicro/ego/core/elog"
	"gorm.io/gorm"
)

type ctxKey string

const (
	bizIDKey    ctxKey = "biz-id"
	uidKey      ctxKey = "uid"
	resourceKey ctxKey = "resource"
)

// gorm校验插件
type GormAccessPlugin struct {
	client          permissionv1.PermissionServiceClient
	statementMap    map[StatementType]string
	logger          *elog.Component
	permissionToken string
}
type GormAccessPluginOption func(*GormAccessPlugin)

func WithStatementMap(statementMap map[StatementType]string) GormAccessPluginOption {
	return func(p *GormAccessPlugin) {
		p.statementMap = statementMap
	}
}

func NewGormAccessPlugin(client permissionv1.PermissionServiceClient, permissionToken string, opts ...GormAccessPluginOption) *GormAccessPlugin {
	plugin := &GormAccessPlugin{
		client: client,
		statementMap: map[StatementType]string{
			SELECT: "read",
			UPDATE: "update",
			DELETE: "delete",
			CREATE: "create",
		},
		logger:          elog.DefaultLogger,
		permissionToken: permissionToken,
	}
	for idx := range opts {
		opts[idx](plugin)
	}
	return plugin
}

// Name 返回插件名称
func (p *GormAccessPlugin) Name() string {
	return "GormAccessPlugin"
}

// Initialize 初始化插件，注册GORM回调
func (p *GormAccessPlugin) Initialize(db *gorm.DB) error {
	// 查询操作
	if err := db.Callback().Query().Before("gorm:query").Register("access:before_query", p.query); err != nil {
		return err
	}
	// 创建操作
	if err := db.Callback().Create().Before("gorm:create").Register("metrics:before_create", p.create); err != nil {
		return err
	}
	// 更新操作
	if err := db.Callback().Update().Before("gorm:update").Register("metrics:before_update", p.update); err != nil {
		return err
	}
	// 删除操作
	if err := db.Callback().Delete().Before("gorm:delete").Register("metrics:before_delete", p.delete); err != nil {
		return err
	}

	return nil
}

func (p *GormAccessPlugin) query(db *gorm.DB) {
	p.accessCheck(SELECT, db)
}

func (p *GormAccessPlugin) update(db *gorm.DB) {
	p.accessCheck(UPDATE, db)
}

func (p *GormAccessPlugin) create(db *gorm.DB) {
	p.accessCheck(CREATE, db)
}

func (p *GormAccessPlugin) delete(db *gorm.DB) {
	p.accessCheck(DELETE, db)
}

func (p *GormAccessPlugin) accessCheck(stmtType StatementType, db *gorm.DB) {
	ctx := db.Statement.Context
	uid, err := getUID(ctx)
	if err != nil {
		_ = db.AddError(fmt.Errorf("获取uid失败 %w", err))
		return
	}
	action, ok := p.statementMap[stmtType]
	if !ok {
		return
	}
	// AuthRequired 实现了这个接口就用这个接口resourceKey进行权限判断
	var key, resourceType string
	if val, ok := db.Statement.Model.(AuthRequired); ok {
		key = val.ResourceKey(ctx)
		resourceType = val.ResourceType(ctx)
	}

	if val, rerr := getResource(ctx); rerr == nil {
		key, resourceType = val.Key, val.Type
	}
	if key != "" {
		ctx = context.WithValue(ctx, "Authorization", p.permissionToken)
		resp, perr := p.client.CheckPermission(ctx, &permissionv1.CheckPermissionRequest{
			Uid: uid,
			Permission: &permissionv1.Permission{
				ResourceKey:  key,
				ResourceType: resourceType,
				Actions:      []string{action},
			},
		})
		if perr != nil {
			_ = db.AddError(fmt.Errorf("权限校验失败 %w", err))
			elog.Error("权限校验失败",
				elog.FieldErr(err),
				elog.Int64("uid", uid),
				elog.String("action", action),
				elog.String("resourceKey", key),
				elog.String("resourceType", resourceType),
			)
			return
		}
		if !resp.Allowed {
			_ = db.AddError(fmt.Errorf("权限校验失败 %w", err))
		}
	}
}

func getUID(ctx context.Context) (int64, error) {
	value := ctx.Value(uidKey)
	if value == nil {
		return 0, fmt.Errorf("uid not found in context")
	}

	// 类型断言校验
	uid, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid uid type, expected int64 got %T", value)
	}

	return uid, nil
}

func getResource(ctx context.Context) (Resource, error) {
	value := ctx.Value(resourceKey)
	if value == nil {
		return Resource{}, fmt.Errorf("resource not found in context")
	}
	res, ok := value.(Resource)
	if !ok {
		return Resource{}, fmt.Errorf("invalid resource type, expected permissionv1.Resource")
	}
	return res, nil
}

type Resource struct {
	Key  string
	Type string
}
