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

	ErrAttributeNotFound error = errors.New("对应属性没找到")

	ErrUnknownOperator = errors.New("未知的比较符")

	ErrUnknownDataType = errors.New("未知的数据类型")

	ErrSupportedSignAlgorithm = errors.New("不支持的签名算法")
	ErrDecodeJWTTokenFailed   = errors.New("JWT令牌解析失败")
	ErrInvalidJWTToken        = errors.New("无效的令牌")

	ErrDatabaseError = errors.New("数据库错误")
	ErrKeyNotExist   = errors.New("key 不存在")

	ErrToAsync = errors.New("服务崩溃已转异步")
)
