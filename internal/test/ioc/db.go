package ioc

import (
	"context"
	"database/sql"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"sync"
	"time"

	"github.com/ecodeclub/ekit/retry"
	"github.com/ego-component/egorm"
	"github.com/gotomicro/ego/core/econf"
)

func WaitForDBSetup(dsn string) {
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	if err1 := ping(sqlDB); err1 != nil {
		panic(err1)
	}
}

func ping(sqlDB *sql.DB) error {
	const maxInterval = 10 * time.Second
	const maxRetries = 10
	strategy, err := retry.NewExponentialBackoffRetryStrategy(time.Second, maxInterval, maxRetries)
	if err != nil {
		return err
	}

	const timeout = 5 * time.Second
	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		err = sqlDB.PingContext(ctx)
		cancel()
		if err == nil {
			break
		}
		next, ok := strategy.Next()
		if !ok {
			panic("Ping DB 重试失败......")
		}
		time.Sleep(next)
	}
	return nil
}

var (
	db         *egorm.Component
	initDBOnce sync.Once
)

func InitDBAndTables() *egorm.Component {
	return InitDBAndTablesWithConfig(map[string]any{
		"dsn":   "root:root@tcp(localhost:13316)/permission?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=1s&readTimeout=3s&writeTimeout=3s&multiStatements=true",
		"debug": true,
	})
}

func InitDBAndTablesWithConfig(mysqlConfig map[string]any) *egorm.Component {
	initDBOnce.Do(func() {
		if db != nil {
			return
		}
		econf.Set("mysql", mysqlConfig)
		WaitForDBSetup(econf.GetStringMapString("mysql")["dsn"])
		db = egorm.Load("mysql").Build()
		time.Sleep(2 * time.Second)
		if err := dao.InitTables(db); err != nil {
			panic(err)
		}
	})
	return db
}
