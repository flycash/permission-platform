//go:build manual

package internal

import (
	"log/slog"
	"os"
	"testing"
	"time"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"gitee.com/flycash/permission-platform/pkg/permission/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	groupName = "test"
)

/*
1. 在goland中使用 Run xxx With Coverage 启动 TestNode2，因为其中有time.Sleep所以会卡住
2. 此时立即再次使用 Run xxx With Coverage 启动 TestNode1，因为其中的grpc客户端都传递的是null，所以要通过测试只能通过group cache分布式特性来获取数据
3. 最终的结果是，TestNode1先结束并通过测试；TestNode2后结束也通过测试。
4. 数据获取流程：
   1. TestNode2 启动并睡眠
   2. TestNode1 启动并立即调用 CheckPermission，先从本地缓存取发现没有，
      group cache在内部利用一致性哈希算法计算得知这个key的所有者应该是 TestNode2，所以向TestNode2发送HTTP请求要数据
   3. TestNode2 虽然扔睡眠，但是内部的HTTP监听协程收到了 TestNode1 的请求，执行groupcache.GetterFunc中的代码——调用rbacClient获取数据，
      并将结果通过HTTP接口返回给 TestNode1，与此同时保留一份数据在TestNode2本地。
   4. TestNode1 收到 TestNode2 返回的数据，保存在本地，继续执行流程，最终验证通过。
   5. TestNode2 结束睡眠继续执行 CheckPermission，先从本地缓存中取，发现本地缓存中已经有数据了（步骤3中保存的）继续执行流程，最终验证通过。
      注意：这里TestNode2中rbacClient被断言了只调用一次Times(1)，足以证明 TestNode2 中调用 CheckPermission 时，group cache直接从本地缓存中获取，而不是再次调用groupcache.GetterFunc
*/

func TestNode2(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// 权限grpc客户端，测试中不会用到仅用来占位
	permissionClient := mocks.NewMockPermissionServiceClient(ctrl)

	// rbac服务grpc客户端，在缓存未命中的时候，要被调用
	rbacClient := mocks.NewMockRBACServiceClient(ctrl)
	rbacClient.EXPECT().GetAllPermissions(gomock.Any(), &permissionv1.GetAllPermissionsRequest{
		BizId:  3,
		UserId: 1,
	}).Return(&permissionv1.GetAllPermissionsResponse{
		UserPermissions: []*permissionv1.UserPermission{
			{
				BizId:            3,
				UserId:           1,
				PermissionId:     2,
				PermissionName:   "test-permission-2",
				ResourceType:     "resource-type-4",
				ResourceKey:      "resource-key-4",
				PermissionAction: "read",
				Effect:           "allow",
			},
			{
				BizId:            3,
				UserId:           1,
				PermissionId:     2,
				PermissionName:   "test-permission-2",
				ResourceType:     "resource-type-4",
				ResourceKey:      "resource-key-4",
				PermissionAction: "write",
				Effect:           "allow",
			},
		},
	}, nil).Times(1)

	// 初始化GroupCachedClient
	c := NewGroupCachedClient(
		permissionClient,
		rbacClient,
		1<<20,
		groupName,
		"http://localhost:7072", // 当前进程监听的http地址
		[]string{"http://localhost:7071", "http://localhost:7072"}, // 分布式缓存集群中节点的完整地址
		slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// 等待15秒，让你有手动启动 TestNode1 测试的机会
	time.Sleep(15 * time.Second)

	// 验证
	permission, err := c.CheckPermission(t.Context(), &permissionv1.CheckPermissionRequest{
		Uid: 1,
		Permission: &permissionv1.Permission{
			Id:           2,
			BizId:        3,
			Name:         "test-permission-2",
			Description:  "",
			ResourceId:   4,
			ResourceType: "resource-type-4",
			ResourceKey:  "resource-key-4",
			Actions:      []string{"read", "write"},
		},
	})
	assert.NoError(t, err)
	assert.True(t, permission.GetAllowed())
}

func TestNode1(t *testing.T) {
	t.Parallel()

	// 初始化GroupCachedClient，其内部不会调用两个grpc客户端所以直接传递null
	c := NewGroupCachedClient(
		nil,
		nil,
		1<<20,
		groupName,
		"http://localhost:7071", // 当前进程监听的http地址
		[]string{"http://localhost:7071", "http://localhost:7072"}, // 分布式缓存集群中节点的完整地址
		slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// 验证
	permission, err := c.CheckPermission(t.Context(), &permissionv1.CheckPermissionRequest{
		Uid: 1,
		Permission: &permissionv1.Permission{
			Id:           2,
			BizId:        3,
			Name:         "test-permission-2",
			Description:  "",
			ResourceId:   4,
			ResourceType: "resource-type-4",
			ResourceKey:  "resource-key-4",
			Actions:      []string{"read", "write"},
		},
	})
	assert.NoError(t, err)
	assert.True(t, permission.GetAllowed())
}
