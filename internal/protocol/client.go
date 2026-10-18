// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package request

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/fool-pigeon/pigeon/internal/project"
	"resty.dev/v3"
)

type Method string

const (
	MethodGet     Method = http.MethodGet
	MethodPost    Method = http.MethodPost
	MethodPut     Method = http.MethodPut
	MethodDelete  Method = http.MethodDelete
	MethodPatch   Method = http.MethodPatch
	MethodHead    Method = http.MethodHead
	MethodOptions Method = http.MethodOptions
	MethodTrace   Method = http.MethodTrace
	MethodConnect Method = http.MethodConnect
)

var (
	ErrEmptyValue   = errors.New("the value is empty")
	ErrInvalidValue = errors.New("the value is invalid")

	hdrUserAgentKey       = http.CanonicalHeaderKey("User-Agent")
	hdrAcceptKey          = http.CanonicalHeaderKey("Accept")
	hdrAcceptEncodingKey  = http.CanonicalHeaderKey("Accept-Encoding")
	hdrContentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	hdrContentLengthKey   = http.CanonicalHeaderKey("Content-Length")
	hdrContentEncodingKey = http.CanonicalHeaderKey("Content-Encoding")
	hdrContentDisposition = http.CanonicalHeaderKey("Content-Disposition")
	hdrAuthorizationKey   = http.CanonicalHeaderKey("Authorization")
	hdrWwwAuthenticateKey = http.CanonicalHeaderKey("WWW-Authenticate")
	hdrRetryAfterKey      = http.CanonicalHeaderKey("Retry-After")
	hdrCookieKey          = http.CanonicalHeaderKey("Cookie")

	plainTextContentType = "text/plain; charset=utf-8"
	jsonContentType      = "application/json"
	xmlContentType       = "application/xml"

	textKey = "text"
	jsonKey = "json"
	xmlKey  = "xml"

	hdrUserAgentValue = "pigeon@" + project.Version

	Methods = []Method{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodOptions, MethodConnect, MethodPatch, MethodTrace}
	// regex pattern
	pathParamRe = regexp.MustCompile(`(?:^|/){([^${}/]+)}`)
)

type (
	Redirect struct {
		MaxCount int      `toml:"max_count" json:"maxCount"`
		Domains  []string `toml:"domains" json:"domains"`
	}
	cDataError struct {
		Msg string
		Err error
	}
)

func (c *cDataError) Error() string {
	return fmt.Sprintf("[http.client_data] %s:%v", c.Msg, c.Err)
}

func (c *cDataError) Unwrap() error {
	return c.Err
}

type Setting struct {
	Timeout string `toml:"timeout" comment:"time>=0, e.g.:'1m30s' time_unit:ns|us|ms|s|m|h" json:"timeout"`
	Retry   Retry  `toml:"retry" json:"retry"`

	Auth Auth `toml:"auth" comment:"token,basic,digest(only-client)" json:"auth"`

	CookieString      string `toml:"cookies" json:"cookieString"`
	UnescapeQuery     bool   `toml:"unescape_query" json:"unescapeQuery"`
	UnescapeJson      bool   `toml:"unescape_json" json:"unescapeJson"`
	ResponseBodyLimit int64  `toml:"response_body_limit" comment:"if limit<=0,it means no limit" json:"responseBodyLimit"`
}

type (
	ClientInstance = resty.Client

	// [Client] is wrap to store client's data into toml file for [resty.Client].
	// The client will be instanced when called [Client.InstancedClient]
	Client struct {
		BaseURL  string `toml:"base_url" json:"baseURL"`
		ProxyURL string `toml:"proxy_url" json:"proxyURL"`

		Path   map[string]string   `toml:"path" json:"path"`
		Query  map[string][]string `toml:"query" json:"query"`
		Header map[string][]string `toml:"header" json:"header"`

		Redirect Redirect `toml:"redirect" json:"redirect"`
		Setting  Setting
	}
)

func NewClientData() *Client {
	return &Client{}
}

// Setters

func (c *Client) SetBaseURL(url string) *Client {
	c.BaseURL = url
	return c
}
func (c *Client) SetProxyURL(url string) *Client {
	c.ProxyURL = url
	return c
}
func (c *Client) SetPathParams(params map[string]string) *Client {
	c.Path = params
	return c
}
func (c *Client) SetQueryParams(params map[string][]string) *Client {
	c.Query = params
	return c
}
func (c *Client) SetHeaders(headers map[string][]string) *Client {
	c.Header = headers
	return c
}
func (c *Client) SetRedirect(redirect Redirect) *Client {
	c.Redirect = redirect
	return c
}
func (c *Client) SetSetting(setting Setting) *Client {
	c.Setting = setting
	return c
}
func (c *Client) SetTimeout(timeStr string) *Client {
	c.Setting.Timeout = timeStr
	return c
}
func (c *Client) SetRetry(retry Retry) *Client {
	c.Setting.Retry = retry
	return c
}
func (c *Client) SetAuth(auth Auth) *Client {
	c.Setting.Auth = auth
	return c
}
func (c *Client) SetCookieString(cookieStr string) *Client {
	c.Setting.CookieString = cookieStr
	return c
}
func (c *Client) SetUnescapeQuery(unescape bool) *Client {
	c.Setting.UnescapeQuery = unescape
	return c
}
func (c *Client) SetUnescapeJson(unescape bool) *Client {
	c.Setting.UnescapeJson = unescape
	return c
}
func (c *Client) SetResponseBodyLimit(limit int64) *Client {
	c.Setting.ResponseBodyLimit = limit
	return c
}

func (c *Client) InstancedClient() (*ClientInstance, error) {
	client := resty.New()

	// allow the payload for GET and DELETE
	client.SetMethodGetAllowPayload(true).SetMethodDeleteAllowPayload(true)

	client.SetHeader(hdrUserAgentKey, hdrUserAgentValue)

	if baseURL := c.BaseURL; baseURL != "" {
		client.SetBaseURL(baseURL)
	}
	if proxyURL := c.ProxyURL; proxyURL != "" {
		client.SetProxy(proxyURL)
	}
	if pathParams := c.Path; pathParams != nil {
		client.SetPathParams(pathParams)
	}
	if queryParams := c.Query; queryParams != nil {
		for name, values := range queryParams {
			for _, value := range values {
				client.SetQueryParam(name, value)
			}
		}
	}
	if headers := c.Header; headers != nil {
		for name, values := range headers {
			for _, value := range values {
				client.SetHeaderVerbatim(name, value)
			}
		}
	}
	if redirect := c.Redirect; redirect.MaxCount != 0 {
		client.SetRedirectPolicy(
			resty.RedirectFlexiblePolicy(redirect.MaxCount),
			resty.RedirectDomainCheckPolicy(redirect.Domains...),
		)
	}
	if err := c.convertSetting(client); err != nil {
		return nil, err
	}

	return client, nil
}

// set settings for client
func (c *Client) convertSetting(client *ClientInstance) error {
	// timeout
	if timeout := c.Setting.Timeout; timeout != "" {
		dur, err := time.ParseDuration(timeout)
		if err != nil {
			return &cDataError{Msg: "timeout", Err: err}
		}
		if dur > 0 {
			client.SetTimeout(dur)
		}
	}
	// resty
	if err := c.convertRetry(client); err != nil {
		return err
	}
	// cookie
	if s := c.Setting.CookieString; s != "" {
		cookies, err := http.ParseCookie(s)
		if err != nil {
			return &cDataError{Msg: "cookies", Err: err}
		}
		client.SetCookies(cookies)
	}
	// unescape query
	client.SetQueryParamsUnescape(c.Setting.UnescapeQuery)
	// unescape json
	client.SetJSONEscapeHTML(!c.Setting.UnescapeJson)
	// response body limit
	client.SetResponseBodyLimit(c.Setting.ResponseBodyLimit)
	return nil
}

// set retry for client
func (c *Client) convertRetry(client *ClientInstance) error {
	retry := c.Setting.Retry
	if count := retry.Count; count > 0 {
		client.SetRetryCount(count)
	}
	if wtStr := retry.WaitTime; wtStr != "" {
		wt, err := time.ParseDuration(wtStr)
		if err != nil {
			return &cDataError{Msg: "retry.wait_time", Err: err}
		}
		if wt > 0 {
			client.SetRetryWaitTime(wt)
		}
	}
	if mwtStr := retry.MaxWaitTime; mwtStr != "" {
		mwt, err := time.ParseDuration(mwtStr)
		if err != nil {
			return &cDataError{Msg: "retry.max_wait_time", Err: err}
		}
		if mwt > 0 {
			client.SetRetryMaxWaitTime(mwt)
		}
	}
	if isAllow := retry.IsAllowNonIdempotent; isAllow {
		client.SetRetryAllowNonIdempotent(isAllow)
	}
	return nil
}
