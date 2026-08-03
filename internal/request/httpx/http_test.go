package httpx_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/homoarea/pigeon/internal/request"
	"github.com/homoarea/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpRequestType(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(1, 1, "should be qual")
	req := httpx.Request{
		RequestMeta: request.RequestMeta{Name: "req", Protocol: "http"},
		QueryParams: map[string][]string{"q1": {"v1"}, "q2": {"v2.1", "v2.2"}},
		PathParams:  map[string]string{"k1": "v1"},
		Method:      http.MethodPost,
		Url:         "http://localhost:8080/echo",
		Body: httpx.Body{
			Type: "json",
			Data: map[string]map[string]string{
				"K1": {"k1": "v1"},
				"K2": {"k1": "v1", "k2": "v2"},
			},
		},
		Settings: httpx.Settings{EncodeUrl: false, AutoRedirects: true, MaxRedirects: 5},
		Desc:     "this is desc",
	}
	b, err := toml.Marshal(req)
	fmt.Println(string(b))
	if err != nil {
		panic(err)
	}

	var meta request.RequestMeta
	err = toml.Unmarshal(b, &meta)
	if err != nil {
		panic(err)
	}
	resp, err := req.Send()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)

	assert.Equal("http", meta.Protocol, "protocol is http")
}
func TestRawURL(t *testing.T) {
	require := require.New(t)
	req := httpx.Request{}
	require.Equal("", req.RawURL())
	req.Url = "http://localhost"
	require.Equal(req.Url, req.RawURL())

	rawURL := "localhost:8888/{mm}/id?h=114&x=y"
	if err := req.ParseRawURL(rawURL); err != nil {
		t.Fatal(err)
	}
	require.Equal("http://"+rawURL, req.RawURL())

	err := req.ParseRawURL("{{{/")
	fmt.Println(err.Error())
	require.Error(err)
}
func TestResty(t *testing.T) {

}
