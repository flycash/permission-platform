//go:build e2e

package rbac

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/auth"
	"gitee.com/flycash/permission-platform/internal/pkg/jwt"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
	rbacioc "gitee.com/flycash/permission-platform/internal/test/integration/ioc/rbac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
)

// TestMain 测试主函数，设置与清理测试环境
func TestMain(m *testing.M) {
	// 初始化数据库
	_ = testioc.InitDBAndTables()

	// 初始化RBAC服务
	svc := rbacioc.Init()

	// 清理除预设数据外的测试数据
	ctx := context.Background()
	t := &testing.T{}
	cleanTestEnvironment(t, ctx, svc)

	// 运行测试
	exitCode := m.Run()

	// 测试完成后再次清理
	cleanTestEnvironment(t, ctx, svc)

	os.Exit(exitCode)
}

func TestRBACInit(t *testing.T) {
	t.Skip("用于演示生成权限平台的SQL脚本的过程")
	// 1. 先将项目根目录下scripts/mysql/data.sql删掉,用scripts/mysql/database.sql中的内容覆盖scripts/mysql/init.sql中的内容。
	// 2. 执行 make e2e_down && make e2e_up
	// 3. 注释掉t.Skip()，手动执行当前测试
	// 4. 观察 scripts/mysql/init.sql 变化
	// 5. 开启 t.Skip()
	repo := rbacioc.Init().Repo

	dir, jwtAuthKey, jwtIssuer, err := getJWTConfig()
	require.NoError(t, err)
	require.NotEmpty(t, jwtAuthKey)

	jwtToken := jwt.New(jwtAuthKey, jwtIssuer)
	bizID := int64(1)
	svc := rbac.NewInitService(bizID, 999, 3000, jwtToken, repo)

	// 执行
	err = svc.Init(t.Context())
	assert.NoError(t, err)

	// 验证
	bizConfig, err := repo.BusinessConfig().FindByID(t.Context(), bizID)
	assert.NoError(t, err)

	mapClaims, err := jwtToken.Decode(bizConfig.Token)
	assert.NoError(t, err)
	assert.Equal(t, float64(bizID), mapClaims[auth.BizIDName])

	// 调用make命令生成SQL脚本
	cmd := exec.Command("make", "mysqldump")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assert.NoError(t, cmd.Run(), cmd.Dir)
}

func getJWTConfig() (dir, key, issuer string, err error) {
	dir, err = os.Getwd()
	if err != nil {
		return "", "", "", err
	}
	rootDir := filepath.Clean(dir + "/../../../..")
	path := rootDir + "/config/config.yaml"
	f, err := os.Open(path)
	if err != nil {
		return "", "", "", err
	}

	err = econf.LoadFromReader(f, yaml.Unmarshal)
	if err != nil {
		return "", "", "", err
	}
	return rootDir, econf.GetString("jwt.key"), econf.GetString("jwt.issuer"), nil
}
