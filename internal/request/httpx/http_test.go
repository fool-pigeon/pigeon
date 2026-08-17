package httpx_test

import (
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
		Desc: "this is desc",
	}
	b, err := toml.Marshal(req)
	t.Log(string(b))
	t.Logf("%+v\n", req)
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
			t.Logf("%v\n%+v\n\n", req.Body.Type, rp)
		}
	}
	t.Run("send file with the wrong file path", func(t *testing.T) {
		req.Body.Type = httpx.BodyBinary
		req.Body.Data.BinaryPath = "./the_file_not_exist"
		_, err := req.Send()
		require.Error(err)
	})
}

func TestRawURL(t *testing.T) {
	testParse := func(t *testing.T, rawURL string, expected ...string) {
		t.Helper()
		var r httpx.Request
		ans := rawURL
		if len(expected) > 0 {
			ans = expected[0]
		}
		if err := r.ParseRawURL(rawURL); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, ans, r.RawURL())
	}
	testParseErr := func(t *testing.T, rawURL string, expected ...error) {
		t.Helper()
		var r httpx.Request
		err := r.ParseRawURL(rawURL)
		if len(expected) > 0 {
			require.ErrorIs(t, err, expected[0])
		} else {
			require.Error(t, err)
		}
	}

	t.Run("the default", func(t *testing.T) {
		var r httpx.Request
		require.Equal(t, "", r.RawURL())
	})
	t.Run("http normal", func(t *testing.T) {
		testParse(t, "http://localhost")
	})
	t.Run("https normal", func(t *testing.T) {
		testParse(t, "https://example.com")
	})
	t.Run("has no protocol with query params", func(t *testing.T) {
		u := "localhost:8888/mm/id?h=114&x=y"
		testParse(t, u, "http://"+u)
	})
	t.Run("query string has the character needed to URL encoded", func(t *testing.T) {
		testParse(t, "https://localhost:8888/mm/id?h=114&x=y&{=}")
	})
	t.Run("path has the character needed to URL encoded", func(t *testing.T) {
		testParse(t, "https://localhost:8888/{mm}/id?h=114&x=y")
	})
	t.Run("err: empty url", func(t *testing.T) {
		testParseErr(t, "")
	})
	t.Run("parse path params", func(t *testing.T) {
		r := httpx.Request{}
		r.PathParams = map[string]string{"p2": "114"}

		rawURL := "{host}/{p1}/{p2}/{}/${}/{${}}/${{}}/{{x}}/{${x}}/{${x}}"
		r.ParseRawURL(rawURL)
		expected := map[string]string{"host": "", "p1": "", "p2": "114"}
		require.Equal(t, expected, r.PathParams)
	})
}
