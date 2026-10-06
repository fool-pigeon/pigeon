// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package httpx_test

import (
	"testing"

	"github.com/fool-pigeon/pigeon/internal/project"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/fool-pigeon/pigeon/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestInstanceClient(t *testing.T) {
	t.Run("new client data", func(t *testing.T) {
		c, _ := httpx.NewClientData().InstancedClient()
		assert.Equal(t, []string{"pg@" + project.Version}, c.Header()["User-Agent"])
	})
	t.Run("send request with client", func(t *testing.T) {
		server := testutil.NewEchoServer(t)
		defer server.Close()

		c, _ := httpx.NewClientData().SetQueryParams(map[string][]string{"a": {"1"}}).InstancedClient()
		req, _ := httpx.NewReqData("t").SetClient(c).SetURL(server.URL).InstancedRequest()
		res, _ := req.Send()
		assert.Equal(t, "1", testutil.ParseEchoResp(res.Bytes()).Query["a"][0])
	})
}
