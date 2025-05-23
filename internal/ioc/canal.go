package ioc

import (
	"fmt"
	"time"

	"github.com/gotomicro/ego/core/econf"
	"github.com/withlin/canal-go/client"
)

func InitCanalConnector() client.CanalConnector {
	// Config 包含Canal连接的所有配置
	type Config struct {
		Host        string        // Canal服务器主机
		Port        int           // Canal服务器端口
		Username    string        // 用户名
		Password    string        // 密码
		Destination string        // 目标实例名
		Schema      string        // 数据库名
		Table       string        // 表名
		IdleTimeout time.Duration // 空闲超时
		BatchSize   int32         // 每次获取的消息数量
	}

	// DefaultCanalConfig 返回默认配置
	// func DefaultCanalConfig() Config {
	// 	return Config{
	// 	Host:        "canal-server",
	// 	Port:        11111,
	// 	Username:    "canal",
	// 	Password:    "canal_pass",
	// 	Destination: "example",
	// 	Schema:      "permission",
	// 	Table:       "user_roles",
	// 	IdleTimeout: 60 * time.Second,
	// 	BatchSize:   1024,
	// }

	var cfg Config
	err := econf.UnmarshalKey("cache.multilevel", &cfg)
	if err != nil {
		panic(err)
	}

	conn := client.NewSimpleCanalConnector(
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Destination,
		int32(cfg.IdleTimeout/time.Millisecond),
		0, // 客户端ID，0表示自动生成
	)
	// 建立 TCP 连接，不接受context
	if err := conn.Connect(); err != nil {
		panic(err)
	}
	// 订阅表
	// 构造订阅表达式
	tableFilter := fmt.Sprintf("%s\\.%s", cfg.Schema, cfg.Table)
	if err := conn.Subscribe(tableFilter); err != nil {
		panic(fmt.Errorf("订阅失败: %w", err))
	}
	return conn
}
