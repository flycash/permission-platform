package log

import (
	"github.com/gotomicro/ego/core/elog"
	"gorm.io/gorm"
)

type LogPlugin struct {
	logger *elog.Component
}

const (
	UPDATE string = "UPDATE"
	DELETE string = "DELETE"
	CREATE string = "CREATE"
)

// NewGormTracingPlugin 创建一个新的GORM追踪插件
func NewGormLogPlugin() *LogPlugin {
	return &LogPlugin{
		logger: elog.DefaultLogger,
	}
}

// Name 返回插件名称
func (p *LogPlugin) Name() string {
	return "GormLogPlugin"
}

// Initialize 初始化插件，注册GORM回调
func (p *LogPlugin) Initialize(db *gorm.DB) error {
	// 创建操作
	if err := db.Callback().Create().After("gorm:create").Register("log:create:after", p.afterCreate); err != nil {
		return err
	}
	// 更新操作
	if err := db.Callback().Update().After("gorm:update").Register("log:update:after", p.afterUpdate); err != nil {
		return err
	}
	// 删除操作
	if err := db.Callback().Delete().After("gorm:delete").Register("log:delete:after", p.afterDelete); err != nil {
		return err
	}

	return nil
}

func (p *LogPlugin) afterCreate(db *gorm.DB) {
	p.record(CREATE, db)
}

func (p *LogPlugin) afterUpdate(db *gorm.DB) {
	p.record(UPDATE, db)
}

func (p *LogPlugin) afterDelete(db *gorm.DB) {
	p.record(DELETE, db)
}

func (p *LogPlugin) record(typ string, db *gorm.DB) {
	// 获取表名
	tableName := db.Statement.Table
	if tableName == "" {
		tableName = db.Statement.Schema.Table
	}
	// 获取完整的SQL语句
	sql := db.Statement.SQL.String()
	values := db.Statement.Vars

	// 使用GORM的SQL构建器获取完整SQL
	completeSQL := db.Dialector.Explain(sql, values...)

	// 记录日志
	if db.Error != nil {
		p.logger.Error("执行GORM SQL失败",
			elog.String("type", typ),
			elog.String("table", tableName),
			elog.String("sql", completeSQL),
			elog.FieldErr(db.Error),
		)
	} else {
		p.logger.Info("执行GORM SQL成功",
			elog.String("type", typ),
			elog.String("table", tableName),
			elog.String("sql", completeSQL),
		)
	}
}
