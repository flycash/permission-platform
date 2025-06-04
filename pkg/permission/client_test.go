//go:build unit

package permission_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gitee.com/flycash/permission-platform/pkg/permission"
	"github.com/ecodeclub/ekit/spi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestLoadClientSuite(t *testing.T) {
	suite.Run(t, new(LoadClientSuite))
}

type LoadClientSuite struct {
	suite.Suite
}

func (l *LoadClientSuite) SetupTest() {
	t := l.T()
	wd, err := os.Getwd()
	require.NoError(t, err)
	cmd := exec.Command("go", "generate", "./...")
	cmd.Dir = filepath.Clean(filepath.Join(wd, "/../.."))
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, fmt.Sprintf("执行 go generate 失败: %v\n%s", err, output))
}

func (l *LoadClientSuite) Test_LoadClient() {
	t := l.T()
	plugins, err := spi.LoadService[permission.Client]("./testdata/plugins", "PermissionClient")
	assert.NoError(t, err)
	assert.Len(t, plugins, 2)
	assert.ElementsMatch(t, []string{plugins[0].Name(), plugins[1].Name()}, []string{"AuthorizedClient", "AggregateClient"})
}
