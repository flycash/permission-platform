//go:build e2e

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	jwtauth "gitee.com/flycash/permission-platform/internal/api/grpc/interceptor/jwt"
	"gitee.com/flycash/permission-platform/internal/service/rbac"
	rbacioc "gitee.com/flycash/permission-platform/internal/test/integration/ioc/rbac"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
)

func TestRBACInit(t *testing.T) {
	t.Skip("用于演示生成权限平台的SQL脚本的过程")
	// 1. 先将项目根目录下scripts/mysql/data.sql删掉,用scripts/mysql/database.sql中的内容覆盖scripts/mysql/init.sql中的内容。
	// 2. 执行 make e2e_down && make e2e_up
	// 3. 注释掉t.Skip()，手动执行当前测试
	// 4. 观察 scripts/mysql/init.sql 变化
	// 5. 开启 t.Skip()
	repo := rbacioc.Init().Repo

	dir, jwtAuthKey, err := getJWTAuthKey()
	require.NoError(t, err)
	require.NotEmpty(t, jwtAuthKey)

	bizID := int64(1)
	svc := rbac.NewInitService(bizID, 999, 3000, jwtAuthKey, repo)

	// 执行
	err = svc.Init(t.Context())
	assert.NoError(t, err)

	// 验证
	bizConfig, err := repo.BusinessConfig().FindByID(t.Context(), bizID)
	assert.NoError(t, err)
	auth := jwtauth.NewJwtAuth(jwtAuthKey)
	mapClaims, err := auth.Decode(bizConfig.Token)
	assert.NoError(t, err)
	assert.Equal(t, float64(bizID), mapClaims[jwtauth.BizIDName])

	// 调用make命令生成SQL脚本
	cmd := exec.Command("make", "mysqldump")
	cmd.Dir = filepath.Clean(filepath.Join(dir, "..", "..", ".."))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assert.NoError(t, cmd.Run(), cmd.Dir)
}

func getJWTAuthKey() (string, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	name := dir + "/../../../config/config.yaml"
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return "", "", err
	}

	err = econf.LoadFromReader(f, yaml.Unmarshal)
	if err != nil {
		return "", "", err
	}
	return dir, econf.GetString("jwt.key"), nil
}
