package core_test

import (
	"path/filepath"
	"testing"

	"github.com/fool-pigeon/pigeon/internal/core"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("Run completed", func(t *testing.T) {
		tmpDir := t.TempDir()
		fsx.WriteToml(filepath.Join(tmpDir, "r1.toml"), httpx.NewReqData("r1"))
		reqNames := make([]string, 0)
		core.Run(tmpDir, []string{tmpDir}, func(res request.Result) {
			reqNames = append(reqNames, res.ReqName)
		})
		assert.Equal(t, "r1", reqNames[0])
	})
}
