package test

import (
	"testing"

	"github.com/homoarea/pigeon/internal/request"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
)

func TestHttpRequestType(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(1, 1, "should be qual")
	req := request.HttpRequest{
		RequestMeta: request.RequestMeta{Name: "req", Protocol: "http"},
		QueryParams: [][2]string{{"pk", "pv1"}, {"pk", "pv2"}},
		PathParams:  [][2]string{{"pk", "pv1"}, {"pk", "pv2"}},
		Method:      "post",
		Url:         "localhost",
		BodyType:    "json",
		Body: `
{
	key:"value",
	data:{
		url:"localhost",
		arr: [1,2,3]
	}
}
		`,
		Settings: request.HttpSettings{EncodeUrl: false, AutoRedirects: true, MaxRedirects: 5},
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
	assert.Equal("http", meta.Protocol, "protocol is http")
}
