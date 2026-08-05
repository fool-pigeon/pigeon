package httpx_test

import (
	"fmt"
	"testing"

	"github.com/homoarea/pigeon/internal/request"
	"github.com/homoarea/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestHttpRequestType(t *testing.T) {
	require := require.New(t)
	bd := httpx.BodyData{
		Raw:            "this is raw line",
		MultiForm:      []httpx.MultipartField{{Name: "n1", Values: []string{"v1", "v2"}}},
		FormUrlEncoded: map[string][]string{"k1": {"v1:", "v2"}},
		BinaryPath:     "../../../assets/test_file.txt",
	}
	req := httpx.Request{
		RequestMeta: request.RequestMeta{Name: "req", Protocol: "http"},
		QueryParams: map[string][]string{"q1": {"v1"}, "q2": {"v2.1", "v2.2"}},
		PathParams:  map[string]string{"k1": "v1"},
		Method:      httpx.MethodPost,
		Url:         "http://localhost:8080/echo",
		Body: httpx.Body{
			Type: httpx.BodyBinary,
			Data: bd,
		},
		Settings: httpx.Settings{EncodeUrl: false, AutoRedirects: true, MaxRedirects: 5},
		Desc:     "this is desc",
	}
	b, err := toml.Marshal(req)
	fmt.Println(string(b))
	fmt.Printf("%+v\n", req)
	if err != nil {
		t.Fatal(err)
	}
	var meta request.RequestMeta
	if err := toml.Unmarshal(b, &meta); err != nil {
		t.Fatal(err)
	}
	require.Equal("http", meta.Protocol, "protocol is http")
	{
		bts := []string{httpx.BodyRaw, httpx.BodyMultiForm, httpx.BodyFormUrlEncoded, httpx.BodyBinary}
		for _, v := range bts {
			req.Body.Type = v
			rp, _ := req.Send()
			fmt.Printf("%v\n%+v\n\n", req.Body.Type, rp)
		}
	}
	{
		req.Body.Type = httpx.BodyBinary
		req.Body.Data.BinaryPath = "./the_file_not_exist"
		_, err := req.Send()
		require.Error(err)
	}
}

func TestRawURL(t *testing.T) {
	require := require.New(t)
	req := httpx.Request{}
	require.Equal("", req.RawURL())
	// http normal
	{
		u := "http://localhost"
		req.ParseRawURL(u)
		require.Equal(u, req.RawURL())
	}
	// https normal
	{
		u := "https://example.com"
		req.ParseRawURL(u)
		require.Equal(u, req.RawURL())
	}
	// has no protocol with query params
	{
		rawURL := "localhost:8888/mm/id?h=114&x=y"
		err := req.ParseRawURL(rawURL)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal("http://"+rawURL, req.RawURL())
	}
	// query string has the character needed to URL encoded
	{
		rawURL := "https://localhost:8888/mm/id?h=114&x=y&{=}"
		req.ParseRawURL(rawURL)
		require.Equal(rawURL, req.RawURL())
	}
	// path has the character needed to URL encoded
	{
		rawURL := "https://localhost:8888/{mm}/id?h=114&x=y"
		req.ParseRawURL(rawURL)
		require.Equal(rawURL, req.RawURL())
	}
	// has username field
	{
		rawURL := "https://user@localhost:8888/{mm}/id?h=114&x=y"
		req.ParseRawURL(rawURL)
		require.Equal(rawURL, req.RawURL())
	}
	// has password field
	{
		rawURL := "https://user:pwd@localhost:8888/{mm}/id?h=114&x=y"
		req.ParseRawURL(rawURL)
		require.Equal(rawURL, req.RawURL())
	}
	// username has char need to URL encoded,cause error
	{
		rawURL := "https://use{}r:pwd@localhost:8888/{mm}/id?h=114&x=y"
		err := req.ParseRawURL(rawURL)
		require.Error(err)
	}
	{
		err := req.ParseRawURL("{{{/")
		require.Error(err)
	}
	// err:rawURL is empty.
	{
		rawURL := ""
		err := req.ParseRawURL(rawURL)
		require.Error(err)
	}
	// err:host is empty.
	{
		err := req.ParseRawURL("http://")
		require.Error(err)
	}
	//TODO: test parse path params
	{

	}
}
