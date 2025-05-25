package ginx

import (
	"context"
	"net/http"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"

	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
)

type CheckPermissionMiddlewareBuilder struct {
	sp              session.Provider
	svc             permissionv1.PermissionServiceClient
	logger          *elog.Component
	permissionToken string
}

func NewCheckPermissionMiddlewareBuilder(svc permissionv1.PermissionServiceClient, token string) *CheckPermissionMiddlewareBuilder {
	return &CheckPermissionMiddlewareBuilder{
		svc:             svc,
		logger:          elog.DefaultLogger,
		permissionToken: token,
		sp:              session.DefaultProvider(),
	}
}

func (c *CheckPermissionMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gCtx := &ginx.Context{Context: ctx}
		sess, err := c.sp.Get(gCtx)
		if err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			c.logger.Debug("用户未登录", elog.FieldErr(err))
			return
		}

		uid := sess.Claims().Uid
		req := &permissionv1.CheckPermissionRequest{
			Uid: uid,
			Permission: &permissionv1.Permission{
				ResourceKey:  gCtx.Request.URL.Path,
				ResourceType: "api",
				Actions:      []string{gCtx.Request.Method},
			},
		}
		// 实时校验权限（慢路径）
		pCtx := context.WithValue(ctx.Request.Context(), "Authorization", c.permissionToken)
		resp, err := c.svc.CheckPermission(pCtx, req)
		if err != nil || !resp.Allowed {
			gCtx.AbortWithStatus(http.StatusForbidden)
			c.logger.Debug("用户无权限", elog.FieldErr(err))
			return
		}
	}
}
