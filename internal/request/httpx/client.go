package httpx

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"resty.dev/v3"
)

const (
	defaultUserAgent = "pigeon/http 0.1v"
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

	CommentedClientData struct {
		BaseURL  string `toml:"base_url,commented"`
		ProxyURL string `toml:"proxy_url,commented"`

		PathParams  map[string]string `toml:"path_params,commented"`
		QueryParams url.Values        `toml:"query_params,commented"`
		Headers     http.Header       `toml:"headers,commented"`

		Redirect Redirect `toml:"redirect,commented"`
		Setting  Setting  `toml:"setting,commented"`
	}

	ClientData struct {
		BaseURL  string `toml:"base_url"`
		ProxyURL string `toml:"proxy_url"`

		PathParams  map[string]string `toml:"path_params"`
		QueryParams url.Values        `toml:"query_params"`
		Headers     http.Header       `toml:"headers"`

		Redirect Redirect `toml:"redirect"`
		Setting  Setting  `toml:"setting"`
	}
)

func NewClientData() ClientData {
	return ClientData{}
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
	c.Headers = headers
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

func (c *ClientData) ToCommentedData() CommentedClientData {
	return CommentedClientData(*c)
}

func (c *ClientData) InstancedClient() *Client {
	client := resty.New()
	client.SetHeader("User-Agent", defaultUserAgent)
	// TODO: instanced client
	return client
}

// set redirect for client
func (c *ClientData) convertRedirect(client *Client) {
	if redirect := c.Redirect; redirect.MaxCount != 0 {
		client.SetRedirectPolicy(
			resty.RedirectFlexiblePolicy(redirect.MaxCount),
			resty.RedirectDomainCheckPolicy(redirect.Domains...),
		)
	}
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

	switch auth := c.Setting.Auth; auth.Type {
	default:
		return &cDataError{Msg: "auth_type", Err: ErrInvalidValue}

	case AuthNone:
		// nothing
	case AuthTypeBasic:
		creds := auth.Creds
		client.SetBasicAuth(creds.Username, creds.Password)
	case AuthTypeDigest:
		creds := auth.Creds
		client.SetDigestAuth(creds.Username, creds.Password)
	case AuthTypeToken:
		data := auth.Token
		client.SetAuthToken(data.Token)
		if data.Scheme != "" {
			client.SetAuthScheme(data.Scheme)
		}
		if data.Header != "" {
			client.SetHeaderAuthorizationKey(data.Header)
		}
	}
	return nil
}
