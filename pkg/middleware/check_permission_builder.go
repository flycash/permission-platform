package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/service/abac"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
)

type CheckPermissionMiddlewareBuilder struct {
	sp     session.Provider
	svc    abac.PermissionSvc
	logger *elog.Component
}

func NewCheckPermissionMiddlewareBuilder(svc abac.PermissionSvc) *CheckPermissionMiddlewareBuilder {
	return &CheckPermissionMiddlewareBuilder{
		svc:    svc,
		logger: elog.DefaultLogger,
	}
}

func (c *CheckPermissionMiddlewareBuilder) Build() gin.HandlerFunc {
	if c.sp == nil {
		c.sp = session.DefaultProvider()
	}
	return func(ctx *gin.Context) {
		gCtx := &ginx.Context{Context: ctx}
		sess, err := c.sp.Get(gCtx)
		if err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			c.logger.Debug("用户未登录", elog.FieldErr(err))
			return
		}

		// 获取bizID
		bizIDStr := ctx.GetHeader("X-Biz-ID")
		bizID, err := strconv.ParseInt(bizIDStr, 10, 64)
		if bizIDStr == "" || err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			c.logger.Debug("bizId非法", elog.FieldErr(errors.New("无法获取bizId")))
			return
		}

		uid := sess.Claims().Uid
		resourceType := "api"
		resourceKey := gCtx.Request.URL.Path
		PermissionAction := gCtx.Request.Method

		// 校验权限，快路径
		permissionsKey := fmt.Sprintf("user:%d:permissions:%s", uid, resourceType)
		permissionsCache := sess.Get(ctx.Request.Context(), permissionsKey)
		if permissionsCache.Err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			c.logger.Debug("获取权限列表失败",
				elog.FieldErr(err),
				elog.FieldCustomKeyValue("resourceType", resourceType),
				elog.FieldCustomKeyValue("resourceKey", resourceKey),
				elog.FieldCustomKeyValue("permissionAction", PermissionAction),
			)
			return
		}
		permissionLists := make(map[string][]string)
		err = permissionsCache.JSONScan(&permissionLists)
		if err != nil {
			c.logger.Debug("反序列化权限列表失败",
				elog.FieldErr(err),
				elog.FieldCustomKeyValue("resourceType", resourceType),
				elog.FieldCustomKeyValue("resourceKey", resourceKey),
				elog.FieldCustomKeyValue("permissionAction", PermissionAction),
			)
		}
		permissions, ok := permissionLists[resourceKey]
		if ok {
			_, ok = slice.Find(permissions, func(src string) bool {
				return src == PermissionAction
			})
			if ok {
				return
			}
		}

		// 如果Session中不存在,实时校验权限（慢路径）
		ok, err = c.svc.Check(ctx.Request.Context(), bizID, uid, domain.Permission{
			BizID: bizID,
			Resource: domain.Resource{
				BizID: bizID,
				Type:  resourceType,
				Key:   resourceKey,
			},
			Action: PermissionAction,
		}, domain.PermissionRequest{
			SubjectAttrs:     map[string]string{},
			ResourceAttrs:    map[string]string{},
			EnvironmentAttrs: map[string]string{},
		})
		if err != nil || !ok {
			gCtx.AbortWithStatus(http.StatusForbidden)
			c.logger.Debug("用户无权限", elog.FieldErr(err))
			return
		}

		// 更新Session
		if _, ok = permissionLists[resourceKey]; !ok {
			permissionLists[resourceKey] = make([]string, 0, 1)
		}
		permissionLists[resourceKey] = append(permissionLists[resourceKey], PermissionAction)
		v, _ := json.Marshal(permissionLists)
		err = sess.Set(ctx.Request.Context(), permissionsKey, v)
		if err != nil {
			elog.Error("更新Session失败", elog.Int64("uid", uid), elog.FieldErr(err))
			gCtx.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}
