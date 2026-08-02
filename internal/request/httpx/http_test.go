package httpx_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/homoarea/pigeon/internal/request"
	"github.com/homoarea/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
)

func TestHttpRequestType(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(1, 1, "should be qual")
	req := httpx.Request{
		RequestMeta: request.RequestMeta{Name: "req", Protocol: "http"},
		QueryParams: [][2]string{{"pk", "pv1"}, {"pk", "pv2"}},
		PathParams:  [][2]string{{"pk", "pv1"}, {"pk", "pv2"}},
		Method:      http.MethodPost,
		Url:         "http://localhost:8080/echo",
		BodyType:    "json",
		Body: `
		{
			"key": "value",
			"arr":[1,2,3]
		}
		`,
		Settings: httpx.Settings{EncodeUrl: false, AutoRedirects: true, MaxRedirects: 5},
		Docs:     "this is docs",
	}
	b, err := toml.Marshal(req)
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
