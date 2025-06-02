package internal

import (
	"context"
	"net"
	"testing"
	"time"

	permissionv1 "gitee.com/flycash/permission-platform/api/proto/gen/permission/v1"
	"github.com/gotomicro/ego/core/elog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type MockServer struct {
	permissionv1.UnimplementedBatchPermissionServiceServer
}

func (m *MockServer) BatchCheckPermission(ctx context.Context, request *permissionv1.BatchCheckPermissionRequest) (*permissionv1.BatchCheckPermissionResponse, error) {
	reqs := request.GetRequests()
	ans := make([]bool, 0, len(reqs))
	for _, req := range reqs {
		if req.Uid%2 == 0 {
			ans = append(ans, true)
		} else {
			ans = append(ans, false)
		}
	}
	return &permissionv1.BatchCheckPermissionResponse{
		Allowed: ans,
	}, nil
}

const bufSize = 1024 * 1024

func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestAggregatePermissionClient_CheckPermission(t *testing.T) {
	// 设置内存中的 gRPC 服务器
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	permissionv1.RegisterBatchPermissionServiceServer(s, &MockServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()
	defer s.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer(lis)), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := &AggregatePermissionClient{
		requestsCh:  make(chan *AggregateRequest, 100),
		batchClient: permissionv1.NewBatchPermissionServiceClient(conn),
		logger:      elog.DefaultLogger,
		batchSize:   3,
	}

	// 启动批处理
	go client.StartBatch(ctx)

	// 创建多个测试请求
	testData := []struct {
		uid  int64
		want bool
	}{
		{uid: 2, want: true},
		{uid: 3, want: false},
		{uid: 4, want: true},
		{uid: 5, want: false},
		{uid: 6, want: true},
		{uid: 7, want: false},
		{uid: 8, want: true},
		{uid: 9, want: false},
		{uid: 10, want: true},
		{uid: 11, want: false},
		{uid: 12, want: true},
		{uid: 13, want: false},
		{uid: 14, want: true},
		{uid: 15, want: false},
		{uid: 16, want: true},
		{uid: 17, want: false},
		{uid: 18, want: true},
		{uid: 19, want: false},
		{uid: 20, want: true},
		{uid: 21, want: false},
	}
	var eg errgroup.Group
	for idx := range testData {
		eg.Go(func() error {
			tc := testData[idx]
			reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			req := &permissionv1.CheckPermissionRequest{
				Uid: tc.uid,
			}

			resp, err := client.CheckPermission(reqCtx, req)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, resp.Allowed)
			return nil
		})
	}
	require.NoError(t, eg.Wait())
}
