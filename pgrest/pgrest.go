// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package pgrest

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/oklog/ulid/v2"

	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/protocol"
	pcl "github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/uid"
	"github.com/jaswdr/faker/v2"
	"resty.dev/v3"
)

type (
	Client        = resty.Client
	ActualRequest = resty.Request
	Response      = resty.Response
)
type RESTSender struct {
	client *Client
	// the instanced request
	req *ActualRequest
}

func New(client *resty.Client) *RESTSender {
	return &RESTSender{client: client}
}

// Send implements [protocol.Sender].
func (s *RESTSender) Send(ctx context.Context, req *pcl.Request, resID ulid.ULID, idGen *uid.Atomic64Gen, emitTrace pcl.TraceHandler) (*pcl.Result, error) {
	if req == nil {
		return nil, fmt.Errorf("request:%w", pcl.ErrNilValue)
	}
	reqInstance, err := s.Instance(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := reqInstance.SetResponseDoNotParse(true).Send()
	if err != nil {
		return nil, err
	}
	if strings.Contains(res.Header().Get(pcl.HdrContentTypeKey), "text/event-stream") {
		call := func(event []byte, data []byte) {
			emitTrace(pcl.NewTrace(idGen.Next(), req.ID, resID, pcl.TraceReceived).
				SetMsg(string(event)).SetPayload(data))
		}
		bodyBuf, err := parseSSE(res.Body, call)
		if err != nil {
			return nil, err
		}
		res.Body = bodyBuf
	}
	res.Request.SetResponseDoNotParse(false)

	result := pcl.Result{
		ResMeta: pcl.ResMeta{
			ResultCover: pcl.ResultCover{Success: res.IsStatusSuccess()},
			ID:          resID,
			Status:      res.Status(),
			Duration:    res.Duration(),
			Request:     *req,
		},
		Raw: res,
	}
	if n := len(req.ConfREST.Tests); n > 0 {
		result.Tests = make([]pcl.TestRes, 0, n)
		env := map[string]any{
			"res":  res,
			"fake": faker.New(),
		}
		for name, code := range req.ConfREST.Tests {
			program, err := expr.Compile(code, expr.Env(env), expr.AsBool())
			if err != nil {
				return nil, err
			}
			output, err := expr.Run(program, env)
			if err != nil {
				result.Success = false
				result.Tests = append(result.Tests, pcl.TestRes{
					Name:   name,
					Passed: false,
					ErrStr: err.Error(),
				})
			}

			passed := output.(bool)
			result.Success = passed && result.Success
			result.Tests = append(result.Tests, pcl.TestRes{
				Name:   name,
				Passed: passed,
			})
		}
	}

	return &result, nil
}
func parseSSE(r io.Reader, call func(event []byte, data []byte)) (io.ReadCloser, error) {
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)
	scanner := bufio.NewScanner(tee)
	// TODO config scanner buffer
	// scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var data, event []byte
	for scanner.Scan() {
		line := scanner.Bytes()
		// empty line,means event end
		if len(line) <= 0 {
			if len(data) > 0 {
				if data[len(data)-1] == '\n' {
					data = data[:len(data)-1]
				}
				if len(event) <= 0 {
					event = []byte("message")
				}
				call(event, data)
			}
			data = nil
			event = nil
			continue
		}
		if i := bytes.IndexByte(line, ':'); i >= 0 {
			v := line[i+1:]
			if len(v) > 0 && v[0] == ' ' {
				v = v[1:]
			}
			switch string(line[:i]) {
			case "data":
				data = append(data, v...)
				data = append(data, '\n')
			case "event":
				event = append(event, v...)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return io.NopCloser(&buf), nil
}
func (s *RESTSender) Instance(ctx context.Context, r *pcl.Request) (*ActualRequest, error) {
	if s.client == nil {
		return nil, fmt.Errorf("rest client:%w", protocol.ErrNilValue)
	}
	if r.ConfREST == nil {
		return nil, fmt.Errorf("request config rest:%w", protocol.ErrNilValue)
	}
	if r.Protocol != pcl.ProtocolREST && r.Protocol != pcl.ProtocolGQL {
		return nil, fmt.Errorf("%s:%w", r.Protocol, protocol.ErrNotSupport)
	}

	timeout, err := r.Timeout.Parse()
	if err != nil {
		return nil, err
	}

	req := s.client.R().
		SetURL(r.URL).
		SetMethod(string(r.Method)).
		SetPathParams(r.ConfREST.Path).
		SetQueryParamsFromValues(r.Query).
		SetTimeout(timeout)

	// Filter out empty string keys & set raw headers
	delete(r.Header, "")
	maps.Copy(req.Header, r.Header)

	if err := applyAuth(req, &r.Auth); err != nil {
		return nil, err
	}

	if err := applyBody(req, &r.PayloadREST); err != nil {
		return nil, err
	}

	if err := applyConfigREST(req, r.ConfREST); err != nil {
		return nil, err
	}

	// exec PreReqExpr
	if r.PreReqExpr != "" {
		env := map[string]any{
			// req is a pointer which we can modify its data
			"req":  req,
			"fake": faker.New(),
		}
		code := r.PreReqExpr
		program, err := expr.Compile(code, expr.Env(env))
		if err != nil {
			return nil, err
		}

		if _, err := expr.Run(program, env); err != nil {
			return nil, err
		}
	}

	s.req = req
	return req, nil
}
func applyAuth(req *ActualRequest, auth *pcl.Auth) error {
	switch auth.Kind {
	case pcl.AuthNone:
	case pcl.AuthInherited:

	case pcl.AuthBasic:
		req.SetBasicAuth(auth.Creds.Username, auth.Creds.Password)
	case pcl.AuthToken:
		token := auth.Token
		req.SetAuthToken(token.Value)
		if token.Scheme != "" {
			req.SetAuthScheme(token.Scheme)
		}
		if token.Header != "" {
			req.SetHeaderAuthorizationKey(token.Header)
		}
	}

	return nil
}
func applyConfigREST(req *ActualRequest, cfg *pcl.ConfREST) error {

	if err := applyRetry(req, &cfg.Retry); err != nil {
		return err
	}

	req.SetQueryParamsUnescape(cfg.UnescapeQuery)
	req.SetJSONEscapeHTML(!cfg.UnescapeJSON)
	req.SetResponseBodyLimit(cfg.ResponseBodyLimit)

	return nil
}
func applyRetry(req *ActualRequest, retry *pcl.Retry) error {
	if count := retry.Count; count > 0 {
		req.SetRetryCount(count)
	}
	if wtStr := retry.WaitTime; wtStr != "" {
		wt, err := wtStr.Parse()
		if err != nil {
			return fmt.Errorf("retry wait time:%w", err)
		}
		if wt > 0 {
			req.SetRetryWaitTime(wt)
		}
	}
	if mwtStr := retry.MaxWaitTime; mwtStr != "" {
		mwt, err := mwtStr.Parse()
		if err != nil {
			return fmt.Errorf("retry max wait time:%w", err)
		}
		if mwt > 0 {
			req.SetRetryMaxWaitTime(mwt)
		}
	}
	if isAllow := retry.IsAllowNonIdempotent; isAllow {
		req.SetRetryAllowNonIdempotent(isAllow)
	}
	return nil
}
func applyBody(req *ActualRequest, b *pcl.Body) error {
	switch b.Kind {
	default:
		return fmt.Errorf("not support body kind:%s", b.Kind)

	case pcl.BodyNone:
		// do nothing

	case pcl.BodyRaw, pcl.BodyGQL, pcl.BodyJSON, pcl.BodyXML:
		req.SetBody(b.Raw)

	case pcl.BodyURLEncoded:
		req.SetFormDataFromValues(b.URLEncoded)

	case pcl.BodyBinary:
		fileBytes, err := fsx.ReadFile(b.BinaryPath)
		if err != nil {
			return err
		}
		req.SetBody(fileBytes)

	case pcl.BodyMultipart:
		var buf bytes.Buffer
		mWriter := multipart.NewWriter(&buf)
		for _, fd := range b.Multipart {
			isSetContentType := fd.ContentType != ""
			isSetFile := fd.FilePath != ""
			var part io.Writer
			var err error
			// set part
			switch {
			case isSetContentType:
				h := make(textproto.MIMEHeader)
				h.Set(pcl.HdrContentTypeKey, fd.ContentType)
				if isSetFile {
					h.Set(pcl.HdrContentDisposition, fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fd.Name, fd.FileName))
				} else {
					h.Set(pcl.HdrContentDisposition, fmt.Sprintf(`form-data; name="%s"`, fd.Name))
				}
				if part, err = mWriter.CreatePart(h); err != nil {
					return fmt.Errorf("set part %s failed:%w", fd.Name, err)
				}
			case isSetFile:
				if part, err = mWriter.CreateFormFile(fd.Name, fd.FileName); err != nil {
					return fmt.Errorf("set part %s failed:%w", fd.Name, err)
				}
			default:
				if part, err = mWriter.CreateFormField(fd.Name); err != nil {
					return fmt.Errorf("set part %s failed:%w", fd.Name, err)
				}
			}
			// write content
			if !isSetFile {
				part.Write([]byte(fd.Value))
			} else {
				fileBytes, err := fsx.ReadFile(fd.FilePath)
				if err != nil {
					return fmt.Errorf("read %s failed:%w", fd.FilePath, err)
				}
				part.Write(fileBytes)
			}
		}
		if err := mWriter.Close(); err != nil {
			return fmt.Errorf("close multipart writer failed:%w", err)
		}
		req.SetContentType(mWriter.FormDataContentType()).
			SetBody(buf.Bytes())
	}
	return nil
}
