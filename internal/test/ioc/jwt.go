package ioc

import (
	"gitee.com/flycash/permission-platform/internal/pkg/jwt"
)

func InitJWTToken() *jwt.Token {
	return jwt.New("permission_platform_key", "permission-platform")
}
