package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/fool-pigeon/pigeon/internal/project"
	"resty.dev/v3"
)

const (
	// MethodGet is the HTTP GET method.
	MethodGet = http.MethodGet
	// MethodPost is the HTTP POST method.
	MethodPost = http.MethodPost
	// MethodPut is the HTTP PUT method.
	MethodPut = http.MethodPut
	// MethodDelete is the HTTP DELETE method.
	MethodDelete = http.MethodDelete
	// MethodPatch is the HTTP PATCH method.
	MethodPatch = http.MethodPatch
	// MethodHead is the HTTP HEAD method.
	MethodHead = http.MethodHead
	// MethodOptions is the HTTP OPTIONS method.
	MethodOptions = http.MethodOptions
	// MethodTrace is the HTTP TRACE method.
	MethodTrace = http.MethodTrace
	// MethodConnect is the HTTP CONNECT method.
	MethodConnect = http.MethodConnect
)

var (
	ErrEmptyValue   = errors.New("the value is empty.")
	ErrInvalidValue = errors.New("the value is invalid.")

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

	hdrUserAgentValue = "pg@" + project.Version

	Methods = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodOptions, MethodConnect, MethodPatch, MethodTrace}
	// regex pattern
	pathParamRe = regexp.MustCompile(`(?:^|/){([^${}/]+)}`)
)

type (
	Redirect struct {
		MaxCount int      `toml:"max_count"`
		Domains  []string `toml:"domains"`
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

type (
	Client = resty.Client

	ClientData struct {
		BaseURL  string `toml:"base_url"`
		ProxyURL string `toml:"proxy_url"`

		PathParams  map[string]string `toml:"path"`
		QueryParams url.Values        `toml:"query"`
		Header      http.Header       `toml:"header"`

		Redirect Redirect `toml:"redirect"`
		Setting  Setting  `toml:"setting"`
	}
)

func NewClientData() *ClientData {
	return &ClientData{}
}

// Setters

func (c *ClientData) SetBaseURL(url string) *ClientData {
	c.BaseURL = url
	return c
}
func (c *ClientData) SetProxyURL(url string) *ClientData {
	c.ProxyURL = url
	return c
}
func (c *ClientData) SetPathParams(params map[string]string) *ClientData {
	c.PathParams = params
	return c
}
func (c *ClientData) SetQueryParams(params url.Values) *ClientData {
	c.QueryParams = params
	return c
}
func (c *ClientData) SetHeaders(headers http.Header) *ClientData {
	c.Header = headers
	return c
}
func (c *ClientData) SetRedirect(redirect Redirect) *ClientData {
	c.Redirect = redirect
	return c
}
func (c *ClientData) SetSetting(setting Setting) *ClientData {
	c.Setting = setting
	return c
}
func (c *ClientData) SetTimeout(timeStr string) *ClientData {
	c.Setting.Timeout = timeStr
	return c
}
func (c *ClientData) SetRetry(retry Retry) *ClientData {
	c.Setting.Retry = retry
	return c
}
func (c *ClientData) SetAuth(auth Auth) *ClientData {
	c.Setting.Auth = auth
	return c
}
func (c *ClientData) SetCookieString(cookieStr string) *ClientData {
	c.Setting.CookieString = cookieStr
	return c
}
func (c *ClientData) SetUnescapeQuery(unescape bool) *ClientData {
	c.Setting.UnescapeQuery = unescape
	return c
}
func (c *ClientData) SetUnescapeJson(unescape bool) *ClientData {
	c.Setting.UnescapeJson = unescape
	return c
}
func (c *ClientData) SetResponseBodyLimit(limit int64) *ClientData {
	c.Setting.ResponseBodyLimit = limit
	return c
}

func (c *ClientData) InstancedClient() (*Client, error) {
	client := resty.New()
	client.SetHeader(hdrUserAgentKey, hdrUserAgentValue)

	if baseURL := c.BaseURL; baseURL != "" {
		client.SetBaseURL(baseURL)
	}
	if proxyURL := c.ProxyURL; proxyURL != "" {
		client.SetProxy(proxyURL)
	}
	if pathParams := c.PathParams; pathParams != nil {
		client.SetPathParams(pathParams)
	}
	if queryParams := c.QueryParams; queryParams != nil {
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
func (c *ClientData) convertSetting(client *Client) error {
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
	// auth
	if err := c.convertAuth(client); err != nil {
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
func (c *ClientData) convertRetry(client *Client) error {
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
	if isAllow := retry.IsAllowNonIdempotent; isAllow != false {
		client.SetRetryAllowNonIdempotent(isAllow)
	}
	return nil
}

// set auth for client
func (c *ClientData) convertAuth(client *Client) error {

	switch auth := c.Setting.Auth; {
	case auth.Baisc != AuthCreds{}:
		creds := auth.Baisc
		client.SetBasicAuth(creds.Username, creds.Password)
	case auth.Digest != AuthCreds{}:
		creds := auth.Digest
		client.SetDigestAuth(creds.Username, creds.Password)
	case auth.Token != AuthToken{}:
		token := auth.Token
		client.SetAuthToken(token.Value)
		if token.Scheme != "" {
			client.SetAuthScheme(token.Scheme)
		}
		if token.Header != "" {
			client.SetHeaderAuthorizationKey(token.Header)
		}
	}
	return nil
}
