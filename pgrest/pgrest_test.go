// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package pgrest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pcl "github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/uid"
	"resty.dev/v3"
)

func newSender(t *testing.T) *RESTSender {
	t.Helper()
	c := resty.New()
	t.Cleanup(func() { c.Close() })
	return New(c)
}

func baseReq(url string) *pcl.Request {
	return &pcl.Request{
		ReqMeta: pcl.ReqMeta{Protocol: pcl.ProtocolREST},
		URL:     url,
		Method:  pcl.MethodGet,
		ReqConf: pcl.ReqConf{
			ConfREST: &pcl.ConfREST{},
		},
	}
}

func TestSend(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		s := newSender(t)
		_, err := s.Send(context.Background(), nil, ulid.Make(), uid.NewU64Gen(), func(*pcl.Trace) {})
		assert.ErrorIs(t, err, pcl.ErrNilValue)
	})

	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		defer srv.Close()

		s := newSender(t)
		res, err := s.Send(context.Background(), baseReq(srv.URL), ulid.Make(), uid.NewU64Gen(), func(*pcl.Trace) {})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.True(t, res.Success)
		assert.True(t, strings.HasPrefix(res.Status, "200"))
	})

	t.Run("failure status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		s := newSender(t)
		res, err := s.Send(context.Background(), baseReq(srv.URL), ulid.Make(), uid.NewU64Gen(), func(*pcl.Trace) {})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.True(t, strings.HasPrefix(res.Status, "500"))
	})

	t.Run("tests expression pass", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"foo":"bar"}`))
		}))
		defer srv.Close()

		req := baseReq(srv.URL)
		req.ConfREST.Tests = map[string]string{
			"status is 200": "res.StatusCode() == 200",
		}
		s := newSender(t)
		res, err := s.Send(context.Background(), req, ulid.Make(), uid.NewU64Gen(), func(*pcl.Trace) {})
		require.NoError(t, err)
		require.Len(t, res.Tests, 1)
		assert.True(t, res.Tests[0].Passed)
		assert.True(t, res.Success)
	})

	t.Run("tests expression fail", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		req := baseReq(srv.URL)
		req.ConfREST.Tests = map[string]string{
			"status is 500": "res.StatusCode() == 500",
		}
		s := newSender(t)
		res, err := s.Send(context.Background(), req, ulid.Make(), uid.NewU64Gen(), func(*pcl.Trace) {})
		require.NoError(t, err)
		require.Len(t, res.Tests, 1)
		assert.False(t, res.Tests[0].Passed)
		assert.False(t, res.Success)
	})

	t.Run("sse emits traces", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			fl, _ := w.(http.Flusher)
			_, _ = w.Write([]byte("event: foo\ndata: bar\n\n"))
			if fl != nil {
				fl.Flush()
			}
		}))
		defer srv.Close()

		var traces []*pcl.Trace
		s := newSender(t)
		_, err := s.Send(context.Background(), baseReq(srv.URL), ulid.Make(), uid.NewU64Gen(), func(tr *pcl.Trace) {
			traces = append(traces, tr)
		})
		require.NoError(t, err)
		require.Len(t, traces, 1)
		assert.Equal(t, "foo", traces[0].Msg)
		assert.Equal(t, "bar", string(traces[0].Payload))
		assert.Equal(t, pcl.TraceReceived, traces[0].Kind)
	})
}

func TestInstance(t *testing.T) {
	t.Run("nil client", func(t *testing.T) {
		s := &RESTSender{}
		_, err := s.Instance(context.Background(), baseReq("http://x"))
		assert.ErrorIs(t, err, pcl.ErrNilValue)
	})

	t.Run("nil config rest", func(t *testing.T) {
		s := newSender(t)
		req := baseReq("http://x")
		req.ConfREST = nil
		_, err := s.Instance(context.Background(), req)
		assert.ErrorIs(t, err, pcl.ErrNilValue)
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		s := newSender(t)
		req := baseReq("http://x")
		req.Protocol = pcl.ProtocolGRPC
		_, err := s.Instance(context.Background(), req)
		assert.ErrorIs(t, err, pcl.ErrNotSupport)
	})

	t.Run("invalid timeout", func(t *testing.T) {
		s := newSender(t)
		req := baseReq("http://x")
		req.Timeout = "bad"
		_, err := s.Instance(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("applies query and headers", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "1", r.URL.Query().Get("foo"))
			assert.Equal(t, "bar", r.Header.Get("X-Test"))
		}))
		defer srv.Close()

		s := newSender(t)
		req := baseReq(srv.URL)
		req.Query = map[string][]string{"foo": {"1"}}
		req.Header = map[string][]string{"X-Test": {"bar"}}
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)

		_, err = actual.Send()
		require.NoError(t, err)
	})

	t.Run("auth basic", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "foo", u)
			assert.Equal(t, "bar", p)
		}))
		defer srv.Close()

		s := newSender(t)
		req := baseReq(srv.URL)
		req.Auth = pcl.Auth{
			Kind:  pcl.AuthBasic,
			Creds: pcl.AuthDataCreds{Username: "foo", Password: "bar"},
		}
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)
		_, err = actual.Send()
		require.NoError(t, err)
	})

	t.Run("auth token custom header", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Token abc", r.Header.Get("X-Auth"))
		}))
		defer srv.Close()

		s := newSender(t)
		req := baseReq(srv.URL)
		req.Auth = pcl.Auth{
			Kind:  pcl.AuthToken,
			Token: pcl.AuthDataToken{Header: "X-Auth", Scheme: "Token", Value: "abc"},
		}
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)
		_, err = actual.Send()
		require.NoError(t, err)
	})

	t.Run("body raw json", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			assert.Equal(t, `{"a":1}`, string(b))
		}))
		defer srv.Close()

		s := newSender(t)
		req := baseReq(srv.URL)
		req.Method = pcl.MethodPost
		req.PayloadREST = pcl.Body{Kind: pcl.BodyJSON, Raw: `{"a":1}`}
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)
		_, err = actual.Send()
		require.NoError(t, err)
	})

	t.Run("body urlencoded", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseForm()
			assert.Equal(t, "1", r.PostFormValue("foo"))
		}))
		defer srv.Close()

		s := newSender(t)
		req := baseReq(srv.URL)
		req.Method = pcl.MethodPost
		req.PayloadREST = pcl.Body{
			Kind:       pcl.BodyURLEncoded,
			URLEncoded: map[string][]string{"foo": {"1"}},
		}
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)
		_, err = actual.Send()
		require.NoError(t, err)
	})

	t.Run("body binary", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "f.bin")
		require.NoError(t, os.WriteFile(p, []byte("xyz"), 0o644))

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			assert.Equal(t, "xyz", string(b))
		}))
		defer srv.Close()

		s := newSender(t)
		req := baseReq(srv.URL)
		req.Method = pcl.MethodPost
		req.PayloadREST = pcl.Body{Kind: pcl.BodyBinary, BinaryPath: p}
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)
		_, err = actual.Send()
		require.NoError(t, err)
	})

	t.Run("body multipart", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
		}))
		defer srv.Close()

		s := newSender(t)
		req := baseReq(srv.URL)
		req.Method = pcl.MethodPost
		req.PayloadREST = pcl.Body{
			Kind: pcl.BodyMultipart,
			Multipart: []pcl.MultiField{
				{Name: "foo", Value: "bar"},
			},
		}
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)
		_, err = actual.Send()
		require.NoError(t, err)
	})

	t.Run("body unsupported kind", func(t *testing.T) {
		s := newSender(t)
		req := baseReq("http://x")
		req.PayloadREST = pcl.Body{Kind: "unknown"}
		_, err := s.Instance(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("retry config", func(t *testing.T) {
		s := newSender(t)
		req := baseReq("http://x")
		req.ConfREST.Retry = pcl.Retry{
			Count:                2,
			WaitTime:             "10ms",
			MaxWaitTime:          "20ms",
			IsAllowNonIdempotent: true,
		}
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, 2, actual.RetryCount)
		assert.Equal(t, 10*time.Millisecond, actual.RetryWaitTime)
		assert.Equal(t, 20*time.Millisecond, actual.RetryMaxWaitTime)
	})

	t.Run("invalid retry wait time", func(t *testing.T) {
		s := newSender(t)
		req := baseReq("http://x")
		req.ConfREST.Retry.WaitTime = "bad"
		_, err := s.Instance(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("pre req expr mutates header", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "injected", r.Header.Get("X-Pre"))
		}))
		defer srv.Close()

		s := newSender(t)
		req := baseReq(srv.URL)
		req.PreReqExpr = `req.SetHeader("X-Pre", "injected")`
		actual, err := s.Instance(context.Background(), req)
		require.NoError(t, err)
		_, err = actual.Send()
		require.NoError(t, err)
	})
}

func TestParseSSE(t *testing.T) {
	t.Run("default event name", func(t *testing.T) {
		var events [][2]string
		rc, err := parseSSE(
			io.NopCloser(strings.NewReader("data: hello\n\n")),
			func(e, d []byte) { events = append(events, [2]string{string(e), string(d)}) },
		)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
		require.Len(t, events, 1)
		assert.Equal(t, "message", events[0][0])
		assert.Equal(t, "hello", events[0][1])
	})

	t.Run("multi data lines joined", func(t *testing.T) {
		var events [][2]string
		_, err := parseSSE(
			io.NopCloser(strings.NewReader("event: foo\ndata: a\ndata: b\n\n")),
			func(e, d []byte) { events = append(events, [2]string{string(e), string(d)}) },
		)
		require.NoError(t, err)
		require.Len(t, events, 1)
		assert.Equal(t, "foo", events[0][0])
		assert.Equal(t, "a\nb", events[0][1])
	})

	t.Run("no space after colon", func(t *testing.T) {
		var events [][2]string
		_, err := parseSSE(
			io.NopCloser(strings.NewReader("event:foo\ndata:bar\n\n")),
			func(e, d []byte) { events = append(events, [2]string{string(e), string(d)}) },
		)
		require.NoError(t, err)
		require.Len(t, events, 1)
		assert.Equal(t, "foo", events[0][0])
		assert.Equal(t, "bar", events[0][1])
	})

	t.Run("empty block ignored", func(t *testing.T) {
		var count int
		_, err := parseSSE(
			io.NopCloser(strings.NewReader("\n\n\ndata: x\n\n")),
			func(e, d []byte) { count++ },
		)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("buffer preserved", func(t *testing.T) {
		rc, err := parseSSE(
			io.NopCloser(strings.NewReader("data: hi\n\n")),
			func(e, d []byte) {},
		)
		require.NoError(t, err)
		defer rc.Close()
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, "data: hi\n\n", string(b))
	})
}

func TestApplyRetry(t *testing.T) {
	t.Run("zero values noop", func(t *testing.T) {
		c := resty.New()
		defer c.Close()
		req := c.R()
		require.NoError(t, applyRetry(req, &pcl.Retry{}))
		assert.Equal(t, 0, req.RetryCount)
	})
	t.Run("negative wait ignored", func(t *testing.T) {
		c := resty.New()
		defer c.Close()
		req := c.R()
		require.NoError(t, applyRetry(req, &pcl.Retry{WaitTime: "-1s"}))
	})
	t.Run("invalid max wait errors", func(t *testing.T) {
		c := resty.New()
		defer c.Close()
		req := c.R()
		assert.Error(t, applyRetry(req, &pcl.Retry{MaxWaitTime: "bad"}))
	})
}

func TestApplyBodyErrors(t *testing.T) {
	t.Run("binary missing file", func(t *testing.T) {
		c := resty.New()
		defer c.Close()
		req := c.R()
		err := applyBody(req, &pcl.Body{Kind: pcl.BodyBinary, BinaryPath: "/nonexistent/x"})
		assert.Error(t, err)
	})
	t.Run("multipart missing file", func(t *testing.T) {
		c := resty.New()
		defer c.Close()
		req := c.R()
		err := applyBody(req, &pcl.Body{
			Kind: pcl.BodyMultipart,
			Multipart: []pcl.MultiField{
				{Name: "f", FilePath: "/nonexistent/x"},
			},
		})
		assert.Error(t, err)
	})
}
