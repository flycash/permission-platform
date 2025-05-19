package ioc

import (
	"gitee.com/flycash/permission-platform/internal/pkg/jwt"
	"github.com/gotomicro/ego/core/econf"
)

func InitJWTToken() *jwt.Token {
	type Config struct {
		Key    string `yaml:"key"`
		Issuer string `yaml:"issuer"`
	}
	var cfg Config
	err := econf.UnmarshalKey("jwt", &cfg)
	if err != nil {
		panic(err)
	}
	return jwt.New(cfg.Key, cfg.Issuer)
}
