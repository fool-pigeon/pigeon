// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package pgrest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/textproto"

	"github.com/expr-lang/expr"

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
func (s *RESTSender) Send(ctx context.Context, req *pcl.Request, idGen *uid.AtomicGen, emitTrace pcl.TraceHandler) (*pcl.Result, error) {
	panic("unimplemented")
}
func (s *RESTSender) Instance(ctx context.Context, r *pcl.Request) (*ActualRequest, error) {
	if s.client == nil {
		return nil, fmt.Errorf("rest client:%w", protocol.ErrNilValue)
	}
	if r.ConfigREST == nil {
		return nil, fmt.Errorf("request config rest:%w", protocol.ErrNilValue)
	}
	if r.Protocol != pcl.ProtocolREST && r.Protocol != pcl.ProtocolGQL {
		return nil, fmt.Errorf("%s:%w", r.Protocol, protocol.ErrNotSupport)
	}

	req := s.client.R().
		SetURL(r.URL).
		SetMethod(string(r.Method)).
		SetPathParams(r.ConfigREST.Path).
		SetQueryParamsFromValues(r.Query)

	// Filter out empty string keys & set raw headers
	delete(r.Header, "")
	maps.Copy(req.Header, r.Header)

	if err := applyAuth(req, &r.Auth); err != nil {
		return nil, err
	}

	if err := applyBody(req, &r.PayloadREST); err != nil {
		return nil, err
	}

	if err := applyConfigREST(req, r.ConfigREST); err != nil {
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
func applyConfigREST(req *ActualRequest, cfg *pcl.ConfigREST) error {

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
	panic("TODO")
}
