//go:build unit

package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitee.com/flycash/permission-platform/pkg/cache"
	cachemocks "gitee.com/flycash/permission-platform/pkg/cache/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// 辅助断言函数
func assertValueEquals(t *testing.T, expected any, res cache.Value) {
	assert.NoError(t, res.Err)
	assert.Equal(t, expected, res.Val)
}

func TestMultipleClusterCache_Set(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) []*cache.Cluster
		key       string
		val       any
		expireIn  time.Duration
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name: "所有集群写入成功",
			mock: func(ctrl *gomock.Controller) []*cache.Cluster {
				clusters := make([]*cache.Cluster, 0, 2)

				cmd1 := cachemocks.NewMockCmdable(ctrl)
				statusCmd1 := redis.NewStatusCmd(context.Background())
				statusCmd1.SetVal("OK")
				cmd1.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "val1", time.Minute).Return(statusCmd1)
				clusters = append(clusters, cache.NewCluster("cluster1", cmd1))

				cmd2 := cachemocks.NewMockCmdable(ctrl)
				statusCmd2 := redis.NewStatusCmd(context.Background())
				statusCmd2.SetVal("OK")
				cmd2.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "val1", time.Minute).Return(statusCmd2)
				clusters = append(clusters, cache.NewCluster("cluster2", cmd2))

				return clusters
			},
			key:       "key1",
			val:       "val1",
			expireIn:  time.Minute,
			assertErr: assert.NoError,
		},
		{
			name: "部分集群写入失败",
			mock: func(ctrl *gomock.Controller) []*cache.Cluster {
				clusters := make([]*cache.Cluster, 0, 2)
				mockErr := errors.New("写入错误")

				cmd1 := cachemocks.NewMockCmdable(ctrl)
				statusCmd1 := redis.NewStatusCmd(context.Background())
				statusCmd1.SetVal("OK")
				cmd1.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "val1", time.Minute).Return(statusCmd1)
				clusters = append(clusters, cache.NewCluster("cluster1", cmd1))

				cmd2 := cachemocks.NewMockCmdable(ctrl)
				statusCmd2 := redis.NewStatusCmd(context.Background())
				statusCmd2.SetErr(mockErr)
				cmd2.EXPECT().Set(gomock.Any(), "permission-platform:multicluster:key1", "val1", time.Minute).Return(statusCmd2)
				clusters = append(clusters, cache.NewCluster("cluster2", cmd2))

				return clusters
			},
			key:      "key1",
			val:      "val1",
			expireIn: time.Minute,
			assertErr: func(t assert.TestingT, err error, i ...any) bool {
				b := assert.Error(t, err)
				assert.Contains(t, err.Error(), "集群[cluster2]: 写入错误")
				return b
			},
		},
		{
			name: "没有集群",
			mock: func(ctrl *gomock.Controller) []*cache.Cluster {
				return []*cache.Cluster{}
			},
			key:       "key1",
			val:       "val1",
			expireIn:  time.Minute,
			assertErr: assert.NoError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			clusters := tc.mock(ctrl)
			c := cache.NewMultipleClusterCache(clusters)

			err := c.Set(context.Background(), tc.key, tc.val, tc.expireIn)
			tc.assertErr(t, err)
		})
	}
}

func TestMultipleClusterCache_Get(t *testing.T) {
	testCases := []struct {
		name         string
		mock         func(ctrl *gomock.Controller) []*cache.Cluster
		key          string
		assertResult func(t *testing.T, res cache.Value)
	}{
		{
			name: "第一个集群找到",
			mock: func(ctrl *gomock.Controller) []*cache.Cluster {
				clusters := make([]*cache.Cluster, 0, 2)

				stringCmd1 := redis.NewStringCmd(context.Background())
				stringCmd1.SetVal("value1")

				cmd1 := cachemocks.NewMockCmdable(ctrl)
				cmd1.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd1)
				clusters = append(clusters, cache.NewCluster("cluster1", cmd1))

				cmd2 := cachemocks.NewMockCmdable(ctrl)
				cmd2.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd1)
				clusters = append(clusters, cache.NewCluster("cluster2", cmd2))
				return clusters
			},
			key: "key1",
			assertResult: func(t *testing.T, res cache.Value) {
				assertValueEquals(t, "value1", res)
			},
		},
		{
			name: "第一个集群未找到，第二个集群找到",
			mock: func(ctrl *gomock.Controller) []*cache.Cluster {
				clusters := make([]*cache.Cluster, 0, 2)

				cmd1 := cachemocks.NewMockCmdable(ctrl)
				stringCmd1 := redis.NewStringCmd(context.Background())
				stringCmd1.SetErr(errors.New("未知错误"))
				cmd1.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd1)
				clusters = append(clusters, cache.NewCluster("cluster1", cmd1))

				cmd2 := cachemocks.NewMockCmdable(ctrl)
				stringCmd2 := redis.NewStringCmd(context.Background())
				stringCmd2.SetVal("value2")
				cmd2.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd2)
				clusters = append(clusters, cache.NewCluster("cluster2", cmd2))

				return clusters
			},
			key: "key1",
			assertResult: func(t *testing.T, res cache.Value) {
				assertValueEquals(t, "value2", res)
			},
		},
		{
			name: "第一批都未找到，但第三个集群找到",
			mock: func(ctrl *gomock.Controller) []*cache.Cluster {
				clusters := make([]*cache.Cluster, 0, 3)

				cmd1 := cachemocks.NewMockCmdable(ctrl)
				stringCmd1 := redis.NewStringCmd(context.Background())
				stringCmd1.SetErr(errors.New("error1"))
				cmd1.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd1)
				clusters = append(clusters, cache.NewCluster("cluster1", cmd1))

				cmd2 := cachemocks.NewMockCmdable(ctrl)
				stringCmd2 := redis.NewStringCmd(context.Background())
				stringCmd2.SetErr(errors.New("error2"))
				cmd2.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd2)
				clusters = append(clusters, cache.NewCluster("cluster2", cmd2))

				cmd3 := cachemocks.NewMockCmdable(ctrl)
				stringCmd3 := redis.NewStringCmd(context.Background())
				stringCmd3.SetVal("value3")
				cmd3.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd3)
				clusters = append(clusters, cache.NewCluster("cluster3", cmd3))

				return clusters
			},
			key: "key1",
			assertResult: func(t *testing.T, res cache.Value) {
				assertValueEquals(t, "value3", res)
			},
		},
		{
			name: "键明确不存在",
			mock: func(ctrl *gomock.Controller) []*cache.Cluster {
				clusters := make([]*cache.Cluster, 0, 1)

				cmd1 := cachemocks.NewMockCmdable(ctrl)
				stringCmd1 := redis.NewStringCmd(context.Background())
				stringCmd1.SetErr(redis.Nil)
				cmd1.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd1)
				clusters = append(clusters, cache.NewCluster("cluster1", cmd1))

				return clusters
			},
			key: "key1",
			assertResult: func(t *testing.T, res cache.Value) {
				assert.True(t, res.KeyNotFound())
			},
		},
		{
			name: "所有集群都返回错误",
			mock: func(ctrl *gomock.Controller) []*cache.Cluster {
				clusters := make([]*cache.Cluster, 0, 2)

				cmd1 := cachemocks.NewMockCmdable(ctrl)
				stringCmd1 := redis.NewStringCmd(context.Background())
				stringCmd1.SetErr(errors.New("error1"))
				cmd1.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd1)
				clusters = append(clusters, cache.NewCluster("cluster1", cmd1))

				cmd2 := cachemocks.NewMockCmdable(ctrl)
				stringCmd2 := redis.NewStringCmd(context.Background())
				stringCmd2.SetErr(errors.New("error2"))
				cmd2.EXPECT().Get(gomock.Any(), "permission-platform:multicluster:key1").Return(stringCmd2)
				clusters = append(clusters, cache.NewCluster("cluster2", cmd2))

				return clusters
			},
			key: "key1",
			assertResult: func(t *testing.T, res cache.Value) {
				assert.Error(t, res.Err)
				assert.Contains(t, res.Err.Error(), "集群[cluster1]: error1")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			clusters := tc.mock(ctrl)
			c := cache.NewMultipleClusterCache(clusters)

			res := c.Get(context.Background(), tc.key)
			tc.assertResult(t, res)
		})
	}
}
