package httpx_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EchoReq struct {
	Method      string              `json:"method"`
	URL         string              `json:"url"`
	Proto       string              `json:"proto"`
	Headers     map[string][]string `json:"headers"`
	Query       map[string][]string `json:"query"`
	Form        map[string][]string `json:"form"`
	Multipart   map[string][]string `json:"multipart"`
	Body        string              `json:"body"`
	BodyBase64  string              `json:"body_base64"`
	ContentType string              `json:"content_type"`
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		_ = r.ParseMultipartForm(32 << 20)
	} else {
		_ = r.ParseForm()
	}

	b, _ := io.ReadAll(r.Body)
	bstr := string(b)
	bBase64 := base64.StdEncoding.EncodeToString(b)

	var multipart map[string][]string
	if form := r.MultipartForm; form != nil {
		multipart = form.Value
	}
	echoReq := EchoReq{
		Method:      r.Method,
		URL:         r.URL.String(),
		Proto:       r.Proto,
		Headers:     r.Header,
		Query:       r.URL.Query(),
		Form:        r.Form,
		Multipart:   multipart,
		Body:        bstr,
		BodyBase64:  bBase64,
		ContentType: r.Header.Get("Content-Type"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(echoReq)

}
func ParseEchoResp(res []byte) EchoReq {
	var e EchoReq
	json.Unmarshal(res, &e)
	return e
}
func NewEchoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(EchoHandler))
}

func TestRun(t *testing.T) {
	server := NewEchoServer(t)
	defer server.Close()

	newReqData := func() *httpx.ReqData {
		return httpx.NewReqData("r").SetURL(server.URL)
	}
	t.Run("only send request", func(t *testing.T) {
		result := newReqData().SetMethod(httpx.MethodPost).SetBodyRaw("abc").Run()
		raw := result.Raw.(*httpx.Response)
		assert.Equal(t, true, result.Success)
		assert.Equal(t, "r", result.ReqName)
		assert.Equal(t, "200 OK", result.Status)
		assert.Nil(t, result.Tests)
		assert.Equal(t, "abc", ParseEchoResp(raw.Bytes()).Body)
		assert.Equal(t, "POST", ParseEchoResp(raw.Bytes()).Method)
		assert.Nil(t, result.Err)
	})
	t.Run("send bad request", func(t *testing.T) {
		result := newReqData().SetURL("http://127.0.0.1:1145141919").Run()
		assert.Equal(t, false, result.Success)
		assert.Equal(t, "r", result.ReqName)
		assert.Error(t, result.Err)
	})
	t.Run("run success with tests expression", func(t *testing.T) {
		tName := "t1"
		tExpr := `res.StatusCode()==200`
		result := newReqData().SetTest(tName, tExpr).Run()
		assert.Equal(t, true, result.Success)

		tRes := result.Tests[0]
		assert.Equal(t, tName, tRes.Name)
		assert.Equal(t, true, tRes.Passed)
		assert.Nil(t, tRes.Err)
	})
	t.Run("run with a expression which will be failed", func(t *testing.T) {
		tName := "t1"
		tExpr := `res.StatusCode()==500`
		result := newReqData().SetTest(tName, tExpr).Run()
		assert.Equal(t, false, result.Success)

		tRes := result.Tests[0]
		assert.Equal(t, tName, tRes.Name)
		assert.Equal(t, false, tRes.Passed)
		assert.Nil(t, tRes.Err)
	})
	t.Run("handle json body in tests", func(t *testing.T) {
		tName := "t1"
		tExpr := `fromJSON(res.String()).method=="GET"`
		result := newReqData().SetTest(tName, tExpr).Run()
		assert.Equal(t, true, result.Success)

		tRes := result.Tests[0]
		assert.Equal(t, true, tRes.Passed)
		assert.Nil(t, tRes.Err)
	})
	t.Run("bad syntax in tests", func(t *testing.T) {
		tName := "t1"
		tExpr := `fromJson(res.String()).method=="GET"`
		result := newReqData().SetTest(tName, tExpr).Run()
		assert.Equal(t, false, result.Success)

		tRes := result.Tests[0]
		assert.Equal(t, false, tRes.Passed)
		assert.Error(t, tRes.Err)
		t.Log(tRes.Err)
	})
}
func TestPreReq(t *testing.T) {
	t.Run("add a header", func(t *testing.T) {
		req, _ := httpx.NewReqData("t").
			SetPreReq(`req.SetHeader("a","2")`).
			InstancedRequest()
		assert.Equal(t, "2", req.Header.Get("a"))
	})
	t.Run("modify the method", func(t *testing.T) {
		req, _ := httpx.NewReqData("t").
			SetPreReq(`req.SetMethod("POST")`).
			InstancedRequest()
		assert.Equal(t, "POST", req.Method)
	})
	t.Run("compile error because bad syntax", func(t *testing.T) {
		_, err := httpx.NewReqData("t").
			SetPreReq(`req.SetMethod("POST"`).
			InstancedRequest()
		t.Log(err)
		assert.Error(t, err)
	})

	t.Run("test faker generator", func(t *testing.T) {
		req, _ := httpx.NewReqData("t").
			SetPreReq(`req.SetBody(faker.Internet().Email())`).
			InstancedRequest()
		assert.Contains(t, req.Body, "@", "faker email should has '@'")
	})
}

func TestReqDataConverts(t *testing.T) {
	t.Run("Convert Body", func(t *testing.T) {
		rData := func(t *testing.T) *httpx.ReqData {
			t.Helper()
			return httpx.NewReqData("t").SetMethod(httpx.MethodPost)
		}
		t.Run("none body", func(t *testing.T) {
			req, _ := rData(t).InstancedRequest()
			assert.Nil(t, req.Body)
		})
		t.Run("raw", func(t *testing.T) {
			req, _ := rData(t).SetBodyRaw("abc").InstancedRequest()
			assert.Equal(t, "abc", req.Body)
		})
		t.Run("JSON String", func(t *testing.T) {
			jstr := `{"a":"1"}`
			req, _ := rData(t).SetBodyRaw(jstr).SetBodyContentType("json").InstancedRequest()
			assert.Equal(t, jstr, req.Body)
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		})
		t.Run("Xml String", func(t *testing.T) {
			xstr := `<?xml version="1.0" encoding="UTF-8"?>
<note>
<to>George</to>
<from>John</from>
<heading>Reminder</heading>
<body>Don't forget the meeting!</body>
</note>`
			req, _ := rData(t).SetBodyRaw(xstr).SetBodyContentType("xml").InstancedRequest()
			assert.Equal(t, xstr, req.Body)
			assert.Equal(t, "application/xml", req.Header.Get("Content-Type"))
		})
		t.Run("multipart/form-data", func(t *testing.T) {
			server := NewEchoServer(t)
			defer server.Close()

			req, _ := rData(t).SetBodyMultipart(httpx.MultiField{Name: "a", Value: "1"}).InstancedRequest()
			res, _ := req.SetURL(server.URL).Send()
			assert.Equal(t, map[string][]string{"a": {"1"}}, ParseEchoResp(res.Bytes()).Multipart)
		})
		t.Run("x-www-url-encoded", func(t *testing.T) {
			req, _ := rData(t).SetBodyForm(map[string][]string{"a": {"1", "2"}}).InstancedRequest()
			assert.Equal(t, url.Values{"a": {"1", "2"}}, req.FormData)
		})
	})
	t.Run("Convert Setting", func(t *testing.T) {
		t.Run("Convert Retry", func(t *testing.T) {
			t.Run("full set", func(t *testing.T) {
				rData := httpx.NewReqData("t").SetRetry(httpx.Retry{
					Count:                3,
					WaitTime:             "100ms",
					MaxWaitTime:          "1s",
					IsAllowNonIdempotent: true,
				})
				req, _ := rData.InstancedRequest()
				assert.Equal(t, 3, req.RetryCount)
				assert.Equal(t, 100*time.Millisecond, req.RetryWaitTime)
				assert.Equal(t, 1*time.Second, req.RetryMaxWaitTime)
				assert.Equal(t, true, req.IsRetryAllowNonIdempotent)
			})
			t.Run("bad format for time", func(t *testing.T) {
				_, err := httpx.NewReqData("t").SetRetry(httpx.Retry{WaitTime: "1kg"}).InstancedRequest()
				assert.Error(t, err)
			})
		})
	})

}

func TestRawURL(t *testing.T) {
	testParse := func(t *testing.T, rawURL string, expected ...string) {
		t.Helper()
		var r httpx.ReqData
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
		var r httpx.ReqData
		err := r.ParseRawURL(rawURL)
		if len(expected) > 0 {
			require.ErrorIs(t, err, expected[0])
		} else {
			require.Error(t, err)
		}
	}

	t.Run("the default", func(t *testing.T) {
		var r httpx.ReqData
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
		r := httpx.ReqData{}
		r.Path = map[string]string{"p2": "114"}

		rawURL := "{host}/{p1}/{p2}/{}/${}/{${}}/${{}}/{{x}}/{${x}}/{${x}}"
		r.ParseRawURL(rawURL)
		expected := map[string]string{"host": "", "p1": "", "p2": "114"}
		require.Equal(t, expected, r.Path)
	})
}
