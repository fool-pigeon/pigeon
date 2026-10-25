// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/fool-pigeon/pigeon/project"
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
	ErrNilValue     = errors.New("value is nil")
	ErrEmptyValue   = errors.New("value is empty")
	ErrInvalidValue = errors.New("the value is invalid")
	ErrNotSupport   = errors.New("the protocol is not support")

	Methods = []Method{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodOptions, MethodConnect, MethodPatch, MethodTrace}

	HdrUserAgentKey       = http.CanonicalHeaderKey("User-Agent")
	HdrAcceptKey          = http.CanonicalHeaderKey("Accept")
	HdrAcceptEncodingKey  = http.CanonicalHeaderKey("Accept-Encoding")
	HdrContentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	HdrContentLengthKey   = http.CanonicalHeaderKey("Content-Length")
	HdrContentEncodingKey = http.CanonicalHeaderKey("Content-Encoding")
	HdrContentDisposition = http.CanonicalHeaderKey("Content-Disposition")
	HdrAuthorizationKey   = http.CanonicalHeaderKey("Authorization")
	HdrWwwAuthenticateKey = http.CanonicalHeaderKey("WWW-Authenticate")
	HdrRetryAfterKey      = http.CanonicalHeaderKey("Retry-After")
	HdrCookieKey          = http.CanonicalHeaderKey("Cookie")

	plainTextContentType = "text/plain; charset=utf-8"
	jsonContentType      = "application/json"
	xmlContentType       = "application/xml"

	textKey = "text"
	jsonKey = "json"
	xmlKey  = "xml"

	hdrUserAgentValue = "pigeon@" + project.Version

	// regex pattern
	pathParamRe = regexp.MustCompile(`(?:^|/){([^${}/]+)}`)
)

type Redirect struct {
	MaxCount int      `json:"maxCount" toml:"max_count"`
	Domains  []string `json:"domains" toml:"domains"`
}

type ActualClient = resty.Client

// [Client] is wrap to store client's data into toml file for [resty.Client].
// The client will be instanced when called [Client.InstancedClient]
type Client struct {
	ProxyURL string      `json:"proxyURL" toml:"proxy_url"`
	Timeout  DurationStr `json:"timeout" toml:"timeout"`
	Redirect Redirect    `json:"redirect" toml:"redirect"`

	client *ActualClient
}

func (c *Client) GetClient() *ActualClient {
	return c.client
}
func (c *Client) InstancedClient() (*ActualClient, error) {
	client := resty.New()
	c.client = client

	// allow the payload for GET and DELETE
	client.SetMethodGetAllowPayload(true).SetMethodDeleteAllowPayload(true)

	client.SetHeader(HdrUserAgentKey, hdrUserAgentValue)

	if proxyURL := c.ProxyURL; proxyURL != "" {
		client.SetProxy(proxyURL)
	}
	if redirect := c.Redirect; redirect.MaxCount != 0 {
		client.SetRedirectPolicy(
			resty.RedirectFlexiblePolicy(redirect.MaxCount),
			resty.RedirectDomainCheckPolicy(redirect.Domains...),
		)
	}

	return client, nil
}
