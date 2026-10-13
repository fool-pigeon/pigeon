// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service_test

import (
	"testing"

	"github.com/fool-pigeon/pigeon/cmd/pg-gui/service"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/fool-pigeon/pigeon/internal/testutil"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
)

func TestSendHttp(t *testing.T) {
	server := testutil.NewEchoServer(t)
	defer server.Close()
	reqCtxService := service.NewReqCtxService(nil)
	reqCtxService.InitCtx("")
	t.Run("Send", func(t *testing.T) {
		reqContent := `name = 'r2'
				protocol = 'http'
				mothod = 'GET'
				url = ''
				`
		var req httpx.ReqData
		toml.Unmarshal([]byte(reqContent), &req)
		req.SetURL(server.URL)
		req.SetHeaders(map[string][]string{"": {""}})
		resHttp, err := reqCtxService.SendHttp(&req)
		assert.Equal(t, server.URL, resHttp.Data.URL)
		assert.Nil(t, err)
	})
}
