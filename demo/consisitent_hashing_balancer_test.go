//go:build unit

package demo

import (
	"context"
	"testing"

	"google.golang.org/grpc/attributes"

	"gitee.com/flycash/permission-platform/pkg/ctxx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type mockSubConn struct {
	balancer.SubConn
	name string
}

func (m *mockSubConn) Name() string {
	return m.name
}

type testMockSubConn struct {
	balancer.SubConn
}

func TestConsistentHashingBalancerBuilder_Build(t *testing.T) {
	// 创建builder
	builder := &ConsistentHashingBalancerBuilder{
		consistentHash: NewConsistentHash(3),
	}

	// 创建测试数据
	readySCs := make(map[balancer.SubConn]base.SubConnInfo)

	// 创建两个测试节点
	subConn1 := &testMockSubConn{}
	subConn2 := &testMockSubConn{}

	// 设置节点信息
	readySCs[subConn1] = base.SubConnInfo{
		Address: resolver.Address{
			Attributes: attributes.New(
				"node", "node1",
			),
		},
	}
	readySCs[subConn2] = base.SubConnInfo{
		Address: resolver.Address{
			Attributes: attributes.New(
				"node", "node2",
			),
		},
	}

	// 构建picker
	picker := builder.Build(base.PickerBuildInfo{
		ReadySCs: readySCs,
	})

	// 验证picker是否正确构建
	assert.NotNil(t, picker)

	// 验证节点是否正确添加
	testBalancer, ok := picker.(*ConsistentHashingBalancer)
	assert.True(t, ok)
	assert.NotNil(t, testBalancer.consistentHash)

	// 验证节点映射
	ctx := context.Background()
	ctx = ctxx.WithBizID(ctx, 1)
	ctx = ctxx.WithUID(ctx, 100)

	result, err := testBalancer.Pick(balancer.PickInfo{Ctx: ctx})
	assert.NoError(t, err)
	assert.NotNil(t, result.SubConn)
}
