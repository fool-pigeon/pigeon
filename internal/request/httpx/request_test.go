package httpx_test

import (
	"testing"

	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/stretchr/testify/require"
)

func TestHttpRequestType(t *testing.T) {

}

func TestRawURL(t *testing.T) {
	testParse := func(t *testing.T, rawURL string, expected ...string) {
		t.Helper()
		var r httpx.RequestData
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
		var r httpx.RequestData
		err := r.ParseRawURL(rawURL)
		if len(expected) > 0 {
			require.ErrorIs(t, err, expected[0])
		} else {
			require.Error(t, err)
		}
	}

	t.Run("the default", func(t *testing.T) {
		var r httpx.RequestData
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
		r := httpx.RequestData{}
		r.PathParams = map[string]string{"p2": "114"}

		rawURL := "{host}/{p1}/{p2}/{}/${}/{${}}/${{}}/{{x}}/{${x}}/{${x}}"
		r.ParseRawURL(rawURL)
		expected := map[string]string{"host": "", "p1": "", "p2": "114"}
		require.Equal(t, expected, r.PathParams)
	})
}
