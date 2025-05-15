package errs

import "errors"

var (
	ErrBizIDNotFound = errors.New("BizID不存在")

	ErrInvalidParameter = errors.New("参数错误")

	ErrRoleDuplicate = errors.New("角色记录biz、type、name唯一索引冲突")

	ErrResourceDuplicate = errors.New("资源记录biz、type、key唯一索引冲突")

	ErrPermissionDuplicate = errors.New("权限记录唯一索引冲突")

	ErrBusinessConfigDuplicate = errors.New("业务配置记录唯一索引冲突")

	ErrRolePermissionDuplicate = errors.New("角色权限关联记录唯一索引冲突")
)
