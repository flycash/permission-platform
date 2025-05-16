//go:build e2e

package rbac

import (
	"fmt"
	"time"

	"gitee.com/flycash/permission-platform/internal/domain"
)

// 定义测试安全常量，避开初始化脚本中的预设数据
const (
	// TestBizID 测试业务ID，避开预设bizID=1
	TestBizID = 2
	// TestUserID 测试用户ID，避开预设userID=999
	TestUserID = 1000
	// TestResourceIDStart 测试资源ID起始值，避开预设资源ID 1-9
	TestResourceIDStart = 10
	// TestPermissionIDStart 测试权限ID起始值，避开预设权限ID 1-18
	TestPermissionIDStart = int64(19)
	// TestRoleIDStart 测试角色ID起始值，避开预设角色ID=1
	TestRoleIDStart = 2
)

// createTestBusinessConfig 创建测试用业务配置对象
func createTestBusinessConfig(name string) domain.BusinessConfig {
	return domain.BusinessConfig{
		Name:      fmt.Sprintf("测试业务-%s", name),
		OwnerID:   TestUserID, // 使用安全的用户ID
		OwnerType: "person",   // 使用有效的拥有者类型：person/organization
		RateLimit: 100,
	}
}

// ActionType 操作类型
type ActionType string

const (
	ActionTypeRead    ActionType = "read"
	ActionTypeWrite   ActionType = "write"
	ActionTypeCreate  ActionType = "create"
	ActionTypeDelete  ActionType = "delete"
	ActionTypeExport  ActionType = "export"
	ActionTypeImport  ActionType = "import"
	ActionTypeExecute ActionType = "execute"
)

// RoleType 角色类型
type RoleType string

const (
	RoleTypeSystem    RoleType = "system"
	RoleTypeCustom    RoleType = "custom"
	RoleTypeTemporary RoleType = "temporary"
)

// createTestRole 创建测试用角色对象
func createTestRole(bizID int64, roleType RoleType) domain.Role {
	now := time.Now().UnixMilli()
	return domain.Role{
		BizID:       bizID,
		Type:        string(roleType),
		Name:        fmt.Sprintf("测试角色-%d", now),
		Description: "测试角色描述",
		Metadata:    fmt.Sprintf(`{"create_time":%d}`, now),
	}
}

// createTestResource 创建测试用资源对象
func createTestResource(bizID int64, resType, key string) domain.Resource {
	now := time.Now().UnixMilli()
	return domain.Resource{
		BizID:       bizID,
		Type:        resType,
		Key:         key,
		Name:        fmt.Sprintf("测试资源-%d", now),
		Description: "测试资源描述",
		Metadata:    fmt.Sprintf(`{"create_time":%d}`, now),
	}
}

// createTestPermission 创建测试用权限对象
func createTestPermission(bizID int64, resource domain.Resource, action ActionType) domain.Permission {
	now := time.Now().UnixMilli()
	return domain.Permission{
		BizID:       bizID,
		Name:        fmt.Sprintf("测试权限-%d", now),
		Description: "测试权限描述",
		Resource:    resource,
		Action:      string(action),
		Metadata:    fmt.Sprintf(`{"create_time":%d}`, now),
	}
}

// createTestUserRole 创建测试用户角色对象
func createTestUserRole(bizID int64, userID int64, role domain.Role) domain.UserRole {
	now := time.Now().UnixMilli()
	return domain.UserRole{
		BizID:     bizID,
		UserID:    userID,
		Role:      role,
		StartTime: now,
		EndTime:   now + 86400000, // 一天后过期
	}
}

// createTestUserPermission 创建测试用户权限对象
func createTestUserPermission(bizID, userID int64, permission domain.Permission, effect domain.Effect) domain.UserPermission {
	now := time.Now().UnixMilli()
	return domain.UserPermission{
		BizID:      bizID,
		UserID:     userID,
		Permission: permission,
		Effect:     effect,
		StartTime:  now,
		EndTime:    now + 86400000, // 一天后过期
	}
}

// createTestRolePermission 创建测试角色权限对象
func createTestRolePermission(bizID int64, role domain.Role, permission domain.Permission) domain.RolePermission {
	return domain.RolePermission{
		BizID:      bizID,
		Role:       role,
		Permission: permission,
	}
}

// createTestRoleInclusion 创建测试角色包含关系对象
func createTestRoleInclusion(bizID int64, includingRole, includedRole domain.Role) domain.RoleInclusion {
	return domain.RoleInclusion{
		BizID:         bizID,
		IncludingRole: includingRole,
		IncludedRole:  includedRole,
	}
}
